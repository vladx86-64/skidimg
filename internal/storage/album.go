package storage

import (
	"context"
	"fmt"
	"skidimg/internal/model"
)

type ImageAlbum struct {
	ImageID int64 `db:"image_id" json:"image_id"`
	AlbumID int64 `db:"album_id" json:"album_id"`
}

// create album
// add to albom
// send list of user albums
// send context of an album (list all images)
// delete from album
// delete album

func (s *Storage) CreateAlbum(ctx context.Context, a *model.Album) (*model.Album, error) {
	query := `
		INSERT INTO albums (user_id, title, description)
		VALUES (:user_id, :title, :description)
		RETURNING id
	`
	stmt, err := s.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error preparing query: %w", err)
	}

	if err := stmt.GetContext(ctx, &a.ID, a); err != nil {
		return nil, fmt.Errorf("error getting inserted ID: %w", err)
	}

	return a, nil
}

func (s *Storage) AddToAlbum(ctx context.Context, a_id int64, imgIds []int64) error {
	links := make([]ImageAlbum, 0, len(imgIds))
	for _, id := range imgIds {
		links = append(links, ImageAlbum{
			ImageID: id,
			AlbumID: a_id,
		})
	}

	query := `
		INSERT INTO image_album (image_id, album_id)
		VALUES (:image_id, :album_id)
	`

	_, err := s.db.NamedExecContext(ctx, query, links)
	if err != nil {
		return fmt.Errorf("failed to insert image album links %w", err)
	}

	return nil
}

func (s *Storage) GetUserAlbums(ctx context.Context, u_id int64) ([]model.Album, error) {
	var albums []model.Album

	query := `SELECT * FROM albums WHERE user_id=$1`
	err := s.db.SelectContext(ctx, &albums, query, u_id)
	if err != nil {
		return nil, err
	}
	return albums, nil
}

func (s *Storage) GetImagesByAlbum(ctx context.Context, a_id int64) ([]model.Image, error) {
	var images []model.Image
	query := `
		SELECT i.* FROM image_album ia
		JOIN images i ON i.id = ia.image_id
		WHERE ia.album_id = $1
	`
	err := s.db.SelectContext(ctx, &images, query, a_id)
	if err != nil {
		return nil, err
	}
	return images, nil
}

func (s *Storage) GetUserImagesNotInAlbum(ctx context.Context, userID, albumID int64) ([]model.Image, error) {
	var images []model.Image
	fmt.Printf("\nuser id: %v album id: %v\n", userID, albumID)

	query := `
		SELECT i.*
		FROM images i
		WHERE i.user_id = $1
		  AND NOT EXISTS (
		    SELECT 1 FROM image_album ia
		    WHERE ia.image_id = i.id AND ia.album_id = $2
		  )
	`

	err := s.db.SelectContext(ctx, &images, query, userID, albumID)
	if err != nil {
		return nil, fmt.Errorf("error selecting available images: %w", err)
	}

	return images, nil
}

func (s *Storage) DeleteImageFromAlbum(ctx context.Context, albumID, imageID int64) error {
	query := `DELETE FROM image_album WHERE album_id = $1 AND image_id = $2`
	_, err := s.db.ExecContext(ctx, query, albumID, imageID)
	if err != nil {
		return fmt.Errorf("error deleting image from album: %w", err)
	}
	return nil
}

func (s *Storage) DeleteAlbum(ctx context.Context, albumID int64) error {
	query := `DELETE FROM albums WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, albumID)
	if err != nil {
		return fmt.Errorf("error deleting album: %w", err)
	}
	return nil
}

// func (s *Storage) DeleteAlbum(ctx context.Context, aid int64) error {
//
// }
