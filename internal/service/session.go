package server

import (
	"context"
	"skidimg/internal/model"
)

func (s *Server) CreateSession(ctx context.Context, session *model.Session) (*model.Session, error) {
	return s.storage.CreateSession(ctx, session)
}

func (s *Server) GetSession(ctx context.Context, id string) (*model.Session, error) {
	return s.storage.GetSession(ctx, id)
}

func (s *Server) RevokeSession(ctx context.Context, id string) error {
	return s.storage.RevokeSession(ctx, id)
}

func (s *Server) DeleteSession(ctx context.Context, id string) error {
	return s.storage.DeleteSession(ctx, id)
}
