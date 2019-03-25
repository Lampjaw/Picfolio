package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/segmentio/ksuid"
)

type Repository struct {
	Database *sql.DB
}

type ImageRecord struct {
	ID          string
	FileType    *string
	Path        string
	Title       *string
	Description *string
	Size        int64
	MimeType    *string
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

func InitRepository() *Repository {
	db, err := sql.Open("sqlite3", "./picfolio.db")
	if err != nil {
		log.Fatal(err)
	}

	sqlStmt := `
		create table images (id text not null primary key, fileType text, title text, path text, description text, size int64, mimeType text, albumId text, height int, width int, created timestamp);
		delete from images;
		create table albums (id text not null primary key, title text, description text, coverPhotoId text, created timestamp);
		delete from albums;
		`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatal("%q: %s\n", err, sqlStmt)
	}

	return &Repository{
		Database: db,
	}
}

func (r *Repository) createAlbumRecord(title string, description string) (string, error) {
	stmt, err := r.Database.Prepare("insert into albums (id, title, description, created) values (?,?,?,?)")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	albumID := generateID()
	now := time.Now().UTC()

	_, err = stmt.Exec(albumID, title, description, now)
	if err != nil {
		return "", err
	}

	return albumID, nil
}

func (r *Repository) createImageRecord(id string, path string, title string, size int64, fileType string, mimeType string, albumID string, height int, width int) error {
	stmt, err := r.Database.Prepare("insert into images (id, path, title, size, fileType, mimeType, albumId, height, width, created) values (?,?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC()

	_, err = stmt.Exec(id, path, title, size, fileType, mimeType, albumID, height, width, now)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) getAllAlbumRecords() ([]*AlbumRecord, error) {
	rows, err := r.Database.Query("select id, title, description, coverPhotoId, created from albums")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records = make([]*AlbumRecord, 0)

	for rows.Next() {
		record := &AlbumRecord{}
		err = rows.Scan(&record.ID, &record.Title, &record.Description, &record.CoverPhotoID, &record.Created)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (r *Repository) getAlbumRecord(id string) (*AlbumRecord, error) {
	stmt, err := r.Database.Prepare("select id, title, description, coverPhotoId, created from albums where id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var record = &AlbumRecord{}
	err = stmt.QueryRow(id).Scan(&record.ID, &record.Title, &record.Description, &record.CoverPhotoID, &record.Created)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (r *Repository) getAllImageRecordsByAlbumID(albumID string) ([]*ImageRecord, error) {
	stmt, err := r.Database.Prepare("select id, path, title, description, size, fileType, mimeType, albumId, height, width, created from images where albumId = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(albumID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records = make([]*ImageRecord, 0)

	for rows.Next() {
		record := &ImageRecord{}
		err = rows.Scan(&record.ID, &record.Path, &record.Title, &record.Description, &record.Size, &record.FileType, &record.MimeType, &record.AlbumID, &record.Height, &record.Width, &record.Created)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (r *Repository) getAllImageRecords() ([]*ImageRecord, error) {
	rows, err := r.Database.Query("select id, path, title, description, size, fileType, mimeType, albumId, height, width, created from images")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records = make([]*ImageRecord, 0)

	for rows.Next() {
		record := &ImageRecord{}
		err = rows.Scan(&record.ID, &record.Path, &record.Title, &record.Description, &record.Size, &record.FileType, &record.MimeType, &record.AlbumID, &record.Height, &record.Width, &record.Created)
		if err != nil {
			return nil, err
		}

		records = append(records, record)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return records, nil
}

func (r *Repository) getImageRecord(id string) (*ImageRecord, error) {
	stmt, err := r.Database.Prepare("select id, path, title, description, size, fileType, mimeType, albumID, height, width, created from images where id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var record = &ImageRecord{}
	err = stmt.QueryRow(id).Scan(&record.ID, &record.Path, &record.Title, &record.Description, &record.Size, &record.FileType, &record.MimeType, &record.AlbumID, &record.Height, &record.Width, &record.Created)
	if err != nil {
		return nil, err
	}

	return record, nil
}

func (r *Repository) setCoverPhotoID(albumID string, coverPhotoID string) error {
	stmt, err := r.Database.Prepare("update albums set coverPhotoId = ? where id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(coverPhotoID, albumID)
	if err != nil {
		return err
	}

	return nil
}

func generateID() string {
	id := ksuid.New()
	return id.String()
}
