package storage

import (
	"context"
	"fmt"
	"skidimg/internal/model"
)

func (s *Storage) CreateImage(ctx context.Context, i *model.Image) (*model.Image, error) {

	query := `
		INSERT INTO images (filename, user_id, orig_ext)
		VALUES (:filename, :user_id, :orig_ext)
		RETURNING id
	`
	stmt, err := s.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error preparing query: %w", err)
	}

	if err := stmt.GetContext(ctx, &i.ID, i); err != nil {
		return nil, fmt.Errorf("error getting inserted ID: %w", err)
	}

	return i, nil

}

func (s *Storage) GetImageByFilename(ctx context.Context, name string) (*model.Image, error) {
	var img model.Image
	query := `SELECT * FROM images WHERE filename=$1`
	err := s.db.GetContext(ctx, &img, query, name)
	if err != nil {
		return nil, err
	}
	return &img, nil
}

func (s *Storage) GetUserGalary(ctx context.Context, u_id int64) ([]model.Image, error) {
	var images []model.Image
	query := `SELECT * FROM images WHERE user_id=$1`
	err := s.db.SelectContext(ctx, &images, query, u_id)
	if err != nil {
		return nil, err
	}
	return images, nil
}
