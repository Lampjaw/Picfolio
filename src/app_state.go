package main

import (
	"os"
	"path/filepath"

	"github.com/segmentio/ksuid"
)

const (
	imageDirectory    = "./data/images/"
	databaseDirectory = "./data/"
	databaseName      = "picfolio.db"
)

type AppState struct {
	exitCallback       chan bool
	imageDirectoryPath string
	databaseFilePath   string
	Repository         *Repository
	AlbumManager       *AlbumManager
	ImageManager       *ImageManager
}

func newAppState() *AppState {
	imageDirectoryPath, _ := filepath.Abs(imageDirectory)
	databaseDirectoryPath, _ := filepath.Abs(databaseDirectory)
	databaseFilePath := filepath.Join(databaseDirectoryPath, databaseName)
	tempImageDirectoryPath := filepath.Join(imageDirectoryPath, "temp")

	os.MkdirAll(imageDirectoryPath, 0755)
	os.MkdirAll(databaseDirectoryPath, 0755)
	os.MkdirAll(tempImageDirectoryPath, 0755)

	state := &AppState{
		exitCallback:       make(chan bool),
		imageDirectoryPath: imageDirectoryPath,
		databaseFilePath:   databaseFilePath,
		Repository:         newRepository(),
	}
	state.AlbumManager = newAlbumManager(state)
	state.ImageManager = newImageManager(state)
	return state
}

func (a AppState) initRepository() {
	a.Repository.initRepository(a.databaseFilePath)
}

func (a AppState) generateID() string {
	id := ksuid.New()
	return id.String()
}
