package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/disintegration/imaging"
)

func deleteAlbum(r *Repository, albumID string) error {
	albumImages, err := r.getAllImageRecordsByAlbumID(albumID)
	if err != nil {
		return err
	}

	for _, image := range albumImages {
		go os.Remove(getThumbFromPath(image.Path))
		go os.Remove(image.Path)
	}

	err = r.deleteImagesByAlbumID(albumID)
	if err != nil {
		return err
	}

	err = r.deleteAlbum(albumID)
	if err != nil {
		return err
	}

	return nil
}

func rotateImage(r *Repository, imageID string) error {
	record, err := r.getImageRecord(imageID)
	if err != nil {
		return err
	}

	if record == nil {
		return nil
	}

	go rotateImageFromPath(record.Path)
	go rotateImageFromPath(getThumbFromPath(record.Path))

	temp := record.Height
	record.Height = record.Width
	record.Width = temp

	err = r.updateImage(imageID, record)
	if err != nil {
		return err
	}

	return nil
}

func rotateImageFromPath(path string) error {
	img, err := imaging.Open(path, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	rotatedImg := imaging.Rotate90(img)
	if err != nil {
		return err
	}

	err = imaging.Save(rotatedImg, path)
	if err != nil {
		return err
	}

	return nil
}

func deleteImage(r *Repository, imageID string) error {
	image, err := r.getImageRecord(imageID)
	if err != nil {
		return err
	}

	album, err := r.getAlbumRecord(image.AlbumID)
	if err != nil {
		return err
	}

	go os.Remove(getThumbFromPath(image.Path))
	go os.Remove(image.Path)

	err = r.deleteImage(imageID)
	if err != nil {
		return err
	}

	if *(album.CoverPhotoID) == imageID {
		images, err := r.getAllImageRecordsByAlbumID(album.ID)
		if err != nil {
			return nil
		}

		if len(images) > 0 {
			err = r.setCoverPhotoID(album.ID, images[0].ID)
			if err != nil {
				return nil
			}
		}
	}

	return nil
}

func getThumbFromPath(path string) string {
	parts := strings.Split(path, ".")
	parts = parts[:len(parts)-1]
	basePath := strings.Join(parts, ".")
	return fmt.Sprintf("%s.thumb.jpg", basePath)
}
