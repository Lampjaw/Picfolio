package main

import "time"

const initSQL = `
CREATE TABLE IF NOT EXISTS images (
	id TEXT NOT NULL PRIMARY KEY,
	fileType TEXT,
	title TEXT,
	path TEXT,
	description TEXT,
	size INT64,
	albumId TEXT,
	height INT,
	width INT,
	created TIMESTAMP
);
CREATE TABLE IF NOT EXISTS albums (
	id TEXT NOT NULL PRIMARY KEY,
	title TEXT,
	description TEXT,
	coverPhotoId TEXT,
	created TIMESTAMP
);
`

type ImageRecord struct {
	ID          string
	FileType    *string
	Path        string
	Title       *string
	Description *string
	Size        int64
	AlbumID     string
	Height      int
	Width       int
	Created     time.Time
}

type AlbumRecord struct {
	ID           string
	Title        string
	Description  *string
	CoverPhotoID *string
	Created      time.Time
}
