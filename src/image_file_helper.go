package main

import (
	"bytes"
	"fmt"
	"image"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

type UploadProfile struct {
	FileType *string
	Path     string
	Title    *string
	Size     int64
	Height   int
	Width    int
}

func newUploadProfile(path string, fileType *string, title *string, size int64, height int, width int) *UploadProfile {
	return &UploadProfile{
		Path:     path,
		FileType: fileType,
		Title:    title,
		Size:     size,
		Height:   height,
		Width:    width,
	}
}

func getEncodingOptions() []imaging.EncodeOption {
	return []imaging.EncodeOption{
		imaging.JPEGQuality(100),
	}
}

func uploadFiles(a *AppState, r *http.Request) ([]*UploadProfile, error) {
	reader, err := r.MultipartReader()
	if err != nil {
		return nil, err
	}

	uploadProfiles := make([]*UploadProfile, 0)

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		if part.FileName() == "" {
			continue
		}

		uploadProfile, err := uploadFile(a, part)
		if err != nil {
			log.Println(err)
			return uploadProfiles, err
		}

		uploadProfiles = append(uploadProfiles, uploadProfile)
	}

	return uploadProfiles, nil
}

func uploadFile(a *AppState, filePart *multipart.Part) (*UploadProfile, error) {
	fileType := getFileType(filePart.FileName())
	fileTitle := filePart.FileName()

	tempFileName := getTempFileName(fileTitle)
	filePath := path.Join(a.imageDirectoryPath, "temp", tempFileName)

	buf := &bytes.Buffer{}
	fileSize, err := io.Copy(buf, filePart)
	if err != nil {
		return nil, err
	}

	img, err := imaging.Decode(buf, imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}

	height := img.Bounds().Dy()
	width := img.Bounds().Dx()

	err = imaging.Save(img, filePath, getEncodingOptions()...)
	if err != nil {
		return nil, err
	}

	uploadProfile := newUploadProfile(filePath, &fileType, &fileTitle, fileSize, height, width)

	return uploadProfile, nil
}

func moveFile(sourcePath string, destPath string) error {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Couldn't open source file: %s", err)
	}

	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("Couldn't open dest file: %s", err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("Writing to output file failed: %s", err)
	}

	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}

	return nil
}

func deleteImage(imagePath string) error {
	if _, err := os.Stat(imagePath); err != nil {
		return nil
	}

	err := os.Remove(imagePath)
	if err != nil {
		return err
	}

	err = deleteThumbnailImage(imagePath)
	if err != nil {
		return err
	}

	return nil
}

func deleteThumbnailImage(imagePath string) error {
	thumbPath := getThumbnailFilePath(imagePath)

	if _, err := os.Stat(thumbPath); err == nil {
		err = os.Remove(thumbPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func rotateImageCCW(imagePath string) error {
	img, err := imaging.Open(imagePath, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	rotatedImg := imaging.Rotate90(img)
	if err != nil {
		return err
	}

	err = imaging.Save(rotatedImg, imagePath, getEncodingOptions()...)
	if err != nil {
		return err
	}

	err = makeThumbnailFromImage(rotatedImg, imagePath)
	if err != nil {
		return err
	}

	return nil
}

func makeThumbnailImage(imagePath string) error {
	img, err := imaging.Open(imagePath)
	if err != nil {
		return err
	}

	err = makeThumbnailFromImage(img, imagePath)
	if err != nil {
		return err
	}

	return nil
}

func makeThumbnailFromImage(img image.Image, imagePath string) error {
	thumbImg := imaging.Fit(img, 650, 650, imaging.Lanczos)
	thumbPath := getThumbnailFilePath(imagePath)

	err := imaging.Save(thumbImg, thumbPath, getEncodingOptions()...)
	if err != nil {
		return err
	}

	return nil
}

func getTempFileName(fileName string) string {
	nameParts := strings.Split(strings.Replace(strings.ToLower(fileName), " ", "-", 0), ".")
	fileType := getFileType(fileName)
	ms := time.Now().UnixNano() / int64(time.Millisecond)
	simpleFileName := strings.Join(nameParts[:len(nameParts)-1], ".")
	return fmt.Sprintf("%s-%d.%s", simpleFileName, ms, fileType)
}

func getThumbnailFilePath(fileName string) string {
	parts := strings.Split(fileName, ".")
	return fmt.Sprintf("%s.thumb.jpg", strings.Join(parts[:len(parts)-1], "."))
}

func getFileType(fileName string) string {
	nameParts := strings.Split(fileName, ".")
	return nameParts[len(nameParts)-1]
}
