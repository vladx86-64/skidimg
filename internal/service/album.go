package server

import (
	"context"
	"skidimg/internal/model"
)

func (s *Server) CreateAlbum(ctx context.Context, a *model.Album) (*model.Album, error) {
	return s.storage.CreateAlbum(ctx, a)
}

func (s *Server) AddToAlbum(ctx context.Context, a_id int64, imgIds []int64) error {
	return s.storage.AddToAlbum(ctx, a_id, imgIds)
}

func (s *Server) GetUserAlbums(ctx context.Context, u_id int64) ([]model.Album, error) {
	return s.storage.GetUserAlbums(ctx, u_id)
}

func (s *Server) GetImagesByAlbum(ctx context.Context, a_id int64) ([]model.Image, error) {
	return s.storage.GetImagesByAlbum(ctx, a_id)
}

func (s *Server) GetUserImagesNotInAlbum(ctx context.Context, u_id, a_id int64) ([]model.Image, error) {
	return s.storage.GetUserImagesNotInAlbum(ctx, u_id, a_id)
}

func (s *Server) DeleteImageFromAlbum(ctx context.Context, a_id, i_id int64) error {
	return s.storage.DeleteImageFromAlbum(ctx, a_id, i_id)
}
