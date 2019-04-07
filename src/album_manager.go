package main

import (
	"os"
	"path"
)

type AlbumManager struct {
	AppState     *AppState
	Repository   *Repository
	ImageManager *ImageManager
}

func newAlbumManager(a *AppState) *AlbumManager {
	return &AlbumManager{
		AppState:     a,
		Repository:   a.Repository,
		ImageManager: a.ImageManager,
	}
}

func (m *AlbumManager) createAlbum(albumName string, albumDescription string) (string, error) {
	albumID := m.AppState.generateID()
	err := m.Repository.createAlbumRecord(albumID, albumName, albumDescription)
	if err != nil {
		return "", err
	}

	albumPath := path.Join(m.AppState.imageDirectoryPath, albumID)

	err = os.MkdirAll(albumPath, 0755)
	if err != nil {
		_ = m.Repository.deleteAlbum(albumID)
		return "", err
	}

	return albumID, err
}

func (m *AlbumManager) getAlbum(albumID string) (*AlbumRecord, error) {
	album, err := m.Repository.getAlbumRecord(albumID)
	if err != nil {
		return nil, err
	}

	return album, nil
}

func (m *AlbumManager) getAllAlbums() ([]*AlbumRecord, error) {
	albums, err := m.Repository.getAllAlbumRecords()
	if err != nil {
		return nil, err
	}

	return albums, nil
}

func (m *AlbumManager) deleteAlbum(albumID string) error {
	albumImages, err := m.Repository.getAllImageRecordsByAlbumID(albumID)
	if err != nil {
		return err
	}

	if len(albumImages) > 0 {
		err = m.Repository.deleteImagesByAlbumID(albumID)
		if err != nil {
			return err
		}
	}

	err = m.Repository.deleteAlbum(albumID)
	if err != nil {
		return err
	}

	err = os.RemoveAll(m.getAlbumPath(albumID))
	if err != nil {
		return err
	}

	return nil
}

func (m *AlbumManager) setAlbumCoverPhoto(albumID string, imageID string) error {
	err := m.Repository.setCoverPhotoID(albumID, imageID)
	if err != nil {
		return err
	}

	return nil
}

func (m *AlbumManager) setAlbumCoverPhotoIfUnset(albumID string, imageID string) error {
	album, err := m.Repository.getAlbumRecord(albumID)
	if err != nil {
		return err
	}

	if album.CoverPhotoID != nil {
		return nil
	}

	err = m.setAlbumCoverPhoto(albumID, imageID)
	if err != nil {
		return err
	}

	return nil
}

func (m *AlbumManager) updateAlbum(albumID string, title string, description *string, coverPhotoID *string) error {
	err := m.Repository.updateAlbum(albumID, title, description, coverPhotoID)
	if err != nil {
		return err
	}

	return nil
}

func (m *AlbumManager) getAlbumPath(albumID string) string {
	return path.Join(m.AppState.imageDirectoryPath, albumID)
}
