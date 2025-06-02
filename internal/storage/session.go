package storage

import (
	"context"
	"fmt"
	"skidimg/internal/model"
)

func (s *Storage) CreateSession(ctx context.Context, session *model.Session) (*model.Session, error) {
	query := `
		INSERT INTO sessions (id, user_email, refresh_token, is_revoked, expires_at)
		VALUES (:id, :user_email, :refresh_token, :is_revoked, :expires_at)
	`
	_, err := s.db.NamedExecContext(ctx, query, session)
	if err != nil {
		return nil, fmt.Errorf("error creating session %w", err)
	}

	return session, nil
}

func (s *Storage) GetSession(ctx context.Context, id string) (*model.Session, error) {
	var session model.Session
	query := `SELECT * FROM sessions WHERE id=$1`
	err := s.db.GetContext(ctx, &session, query, id)
	if err != nil {
		return nil, fmt.Errorf("error getting session %w", err)
	}

	return &session, nil
}

func (s *Storage) RevokeSession(ctx context.Context, id string) error {
	query := `UPDATE sessions SET is_revoked=TRUE WHERE id=$1`
	_, err := s.db.ExecContext(ctx, query, id)

	if err != nil {
		return fmt.Errorf("error revoking session %w", err)
	}

	return nil
}

func (s *Storage) DeleteSession(ctx context.Context, id string) error {
	query := `DELETE FROM sessions WHERE id=$1`
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting session %w", err)
	}

	return nil
}
