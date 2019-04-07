package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Repository struct {
	Database *sql.DB
}

func newRepository() *Repository {
	return &Repository{}
}

func (r *Repository) initRepository(dataSourceName string) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(initSQL)
	if err != nil {
		log.Fatal("%q: %s\n", err, initSQL)
	}

	r.Database = db
}

func (r *Repository) createAlbumRecord(id string, title string, description string) error {
	stmt, err := r.Database.Prepare("insert into albums (id, title, description, created) values (?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC()

	_, err = stmt.Exec(id, title, description, now)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) createImageRecord(id string, path string, title *string, size int64, fileType *string, albumID string, height int, width int) error {
	stmt, err := r.Database.Prepare("insert into images (id, path, title, size, fileType, albumId, height, width, created) values (?,?,?,?,?,?,?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().UTC()

	_, err = stmt.Exec(id, path, title, size, fileType, albumID, height, width, now)
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
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return record, nil
}

func (r *Repository) getAllImageRecordsByAlbumID(albumID string) ([]*ImageRecord, error) {
	stmt, err := r.Database.Prepare("select id, path, title, description, size, fileType, albumId, height, width, created from images where albumId = ?")
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
		err = rows.Scan(&record.ID, &record.Path, &record.Title, &record.Description, &record.Size, &record.FileType, &record.AlbumID, &record.Height, &record.Width, &record.Created)
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
	rows, err := r.Database.Query("select id, path, title, description, size, fileType, albumId, height, width, created from images")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records = make([]*ImageRecord, 0)

	for rows.Next() {
		record := &ImageRecord{}
		err = rows.Scan(&record.ID, &record.Path, &record.Title, &record.Description, &record.Size, &record.FileType, &record.AlbumID, &record.Height, &record.Width, &record.Created)
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
	stmt, err := r.Database.Prepare("select id, path, title, description, size, fileType, albumId, height, width, created from images where id = ?")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var record = &ImageRecord{}
	err = stmt.QueryRow(id).Scan(&record.ID, &record.Path, &record.Title, &record.Description, &record.Size, &record.FileType, &record.AlbumID, &record.Height, &record.Width, &record.Created)
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

func (r *Repository) updateAlbum(albumID string, title string, description *string, coverPhotoID *string) error {
	stmt, err := r.Database.Prepare("update albums set title = ?, description = ?, coverPhotoId = ? where id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(title, description, coverPhotoID, albumID)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) deleteAlbum(albumID string) error {
	stmt, err := r.Database.Prepare("delete from albums where id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(albumID)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) deleteImagesByAlbumID(albumID string) error {
	stmt, err := r.Database.Prepare("delete from images where albumId = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(albumID)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) deleteImage(imageID string) error {
	stmt, err := r.Database.Prepare("delete from images where id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(imageID)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repository) updateImage(imageID string, record *ImageRecord) error {
	stmt, err := r.Database.Prepare("update images set description = ?, height = ?, width = ?, albumId = ? where id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(record.Description, record.Height, record.Width, record.AlbumID, imageID)
	if err != nil {
		return err
	}

	return nil
}
