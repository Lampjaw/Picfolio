package main

import (
	"fmt"
	"image/jpeg"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

const BaseImageDirectory = "./images/"

var imagePath string

func init() {
	absPath, _ := filepath.Abs(BaseImageDirectory)
	os.MkdirAll(absPath, os.ModePerm)
	imagePath = absPath
}

func uploadFiles(repo *Repository, albumID string, r *http.Request) ([]string, error) {
	album, err := repo.getAlbumRecord(albumID)
	if err != nil {
		return nil, err
	}

	reader, err := r.MultipartReader()
	if err != nil {
		return nil, err
	}

	fileIDs := make([]string, 0)

	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}

		if part.FileName() == "" {
			continue
		}

		fileID, err := uploadFile(repo, albumID, part)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		fileIDs = append(fileIDs, fileID)
	}

	if len(fileIDs) > 0 && album.CoverPhotoID == nil {
		_ = repo.setCoverPhotoID(albumID, fileIDs[0])
	}

	return fileIDs, nil
}

func uploadFile(r *Repository, albumID string, filePart *multipart.Part) (string, error) {
	fileID := generateID()

	nameParts := strings.Split(filePart.FileName(), ".")
	fileType := nameParts[len(nameParts)-1]

	fileName := fmt.Sprintf("%s.%s", fileID, fileType)
	filePath := filepath.Join(imagePath, fileName)

	dst, err := os.Create(filePath)
	defer dst.Close()

	if err != nil {
		return "", err
	}

	fileSize, err := io.Copy(dst, filePart)
	if err != nil {
		os.Remove(filePath)
		return "", err
	}

	height, width, err := postUploadActions(filePath)
	if err != nil {
		os.Remove(filePath)
		return "", err
	}

	var mimeType = filePart.Header.Get("Content-Type")

	err = r.createImageRecord(fileID, filePath, filePart.FileName(), fileSize, fileType, mimeType, albumID, height, width)
	if err != nil {
		os.Remove(filePath)
		return "", err
	}

	makeThumbnail(filePath, fileID, fileType)

	return fileID, nil
}

func postUploadActions(filePath string) (int, int, error) {
	img, err := imaging.Open(filePath, imaging.AutoOrientation(true))
	if err != nil {
		return 0, 0, nil
	}

	height := img.Bounds().Dy()
	width := img.Bounds().Dx()

	err = imaging.Save(img, filePath)
	if err != nil {
		return 0, 0, nil
	}

	return height, width, nil
}

func makeThumbnail(filePath string, fileID string, fileType string) error {
	img, err := imaging.Open(filePath, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	m := imaging.Fit(img, 650, 650, imaging.Lanczos)

	thumbName := fmt.Sprintf("%s.thumb.jpg", fileID)
	thumbPath := filepath.Join(imagePath, thumbName)

	dst, err := os.Create(thumbPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	jpeg.Encode(dst, m, nil)

	return nil
}
