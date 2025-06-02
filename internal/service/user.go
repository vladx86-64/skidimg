package server

import (
	"context"
	"skidimg/internal/model"
)

func (s *Server) CreateUser(ctx context.Context, u *model.User) (*model.User, error) {
	return s.storage.CreateUser(ctx, u)
}
func (s *Server) GetUser(ctx context.Context, email string) (*model.User, error) {
	return s.storage.GetUser(ctx, email)
}

func (s *Server) ListUsers(ctx context.Context) ([]model.User, error) {
	return s.storage.ListUsers(ctx)
}

func (s *Server) UpdateUser(ctx context.Context, u *model.User) (*model.User, error) {
	return s.storage.UpdateUser(ctx, u)
}

func (s *Server) DeleteUser(ctx context.Context, id int64) error {
	return s.storage.DeleteUser(ctx, id)
}
