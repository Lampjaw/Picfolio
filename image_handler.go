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

	buf := &bytes.Buffer{}
	fileSize, err := io.Copy(buf, filePart)
	if err != nil {
		return "", err
	}

	img, err := imaging.Decode(buf, imaging.AutoOrientation(true))
	if err != nil {
		return "", err
	}

	height := img.Bounds().Dy()
	width := img.Bounds().Dx()

	err = imaging.Save(img, filePath)
	if err != nil {
		return "", err
	}

	err = makeThumbnail(img, fileID)
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

	return fileID, nil
}

func makeThumbnail(img image.Image, fileID string) error {
	m := imaging.Fit(img, 650, 650, imaging.Lanczos)

	thumbName := fmt.Sprintf("%s.thumb.jpg", fileID)
	thumbPath := filepath.Join(imagePath, thumbName)

	err := imaging.Save(m, thumbPath)
	if err != nil {
		return err
	}

	return nil
}
