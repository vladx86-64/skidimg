package server

import (
	"context"
	"skidimg/internal/model"
)

func (s *Server) CreateImage(ctx context.Context, i *model.Image) (*model.Image, error) {
	return s.storage.CreateImage(ctx, i)
}

func (s *Server) GetImageByFilename(ctx context.Context, name string) (*model.Image, error) {
	return s.storage.GetImageByFilename(ctx, name)
}

func (s *Server) GetUserGalary(ctx context.Context, u_id int64) ([]model.Image, error) {
	return s.storage.GetUserGalary(ctx, u_id)
}
