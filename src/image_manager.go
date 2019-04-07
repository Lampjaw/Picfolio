package main

import (
	"fmt"
	"path"
)

type ImageManager struct {
	AppState     *AppState
	Repository   *Repository
	AlbumManager *AlbumManager
}

func newImageManager(a *AppState) *ImageManager {
	return &ImageManager{
		AppState:     a,
		Repository:   a.Repository,
		AlbumManager: a.AlbumManager,
	}
}

func (m *ImageManager) createImage(albumID string, uploadProfile *UploadProfile) (string, error) {
	imageID := m.AppState.generateID()
	imagePath := m.getImagePath(albumID, imageID, uploadProfile.FileType)

	err := moveFile(uploadProfile.Path, imagePath)
	if err != nil {
		return "", err
	}

	err = m.Repository.createImageRecord(imageID, imagePath, uploadProfile.Title, uploadProfile.Size, uploadProfile.FileType, albumID, uploadProfile.Height, uploadProfile.Width)
	if err != nil {
		return "", err
	}

	err = makeThumbnailImage(imagePath)
	if err != nil {
		return "", err
	}

	err = m.AlbumManager.setAlbumCoverPhotoIfUnset(albumID, imageID)
	if err != nil {
		return "", err
	}

	return imageID, nil
}

func (m *ImageManager) getImage(imageID string) (*ImageRecord, error) {
	image, err := m.Repository.getImageRecord(imageID)
	if err != nil {
		return nil, err
	}

	return image, nil
}

func (m *ImageManager) getAllImagesByAlbumID(albumID string) ([]*ImageRecord, error) {
	images, err := m.Repository.getAllImageRecordsByAlbumID(albumID)
	if err != nil {
		return nil, err
	}

	return images, nil
}

func (m *ImageManager) deleteImage(imageID string) error {
	image, err := m.getImage(imageID)
	if err != nil {
		return err
	}

	album, err := m.AlbumManager.getAlbum(image.AlbumID)
	if err != nil {
		return err
	}

	err = deleteImage(image.Path)
	if err != nil {
		return err
	}

	err = m.Repository.deleteImage(imageID)
	if err != nil {
		return err
	}

	if *(album.CoverPhotoID) == imageID {
		images, err := m.Repository.getAllImageRecordsByAlbumID(album.ID)
		if err != nil {
			return nil
		}

		if len(images) > 0 {
			err = m.AlbumManager.setAlbumCoverPhoto(album.ID, images[0].ID)
			if err != nil {
				return nil
			}
		}
	}

	return nil
}

func (m *ImageManager) updateImage(imageID string, image *ImageRecord) error {
	err := m.Repository.updateImage(imageID, image)
	if err != nil {
		return err
	}

	return nil
}

func (m *ImageManager) rotateImage(imageID string) error {
	record, err := m.Repository.getImageRecord(imageID)
	if err != nil {
		return err
	}

	err = rotateImageCCW(record.Path)
	if err != nil {
		return err
	}

	temp := record.Height
	record.Height = record.Width
	record.Width = temp

	err = m.Repository.updateImage(imageID, record)
	if err != nil {
		return err
	}

	return nil
}

func (m *ImageManager) getImagePath(albumID string, imageID string, fileType *string) string {
	return path.Join(m.AlbumManager.getAlbumPath(albumID), fmt.Sprintf("%s.%s", imageID, *fileType))
}
