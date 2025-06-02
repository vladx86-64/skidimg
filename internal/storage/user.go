package storage

import (
	"context"
	"fmt"
	model "skidimg/internal/model"
)

func (s *Storage) CreateUser(ctx context.Context, u *model.User) (*model.User, error) {

	query := `
		INSERT INTO users (name, email, password, is_admin)
		VALUES (:name, :email, :password, :is_admin)
		RETURNING id
	`
	stmt, err := s.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error preparing query: %w", err)
	}

	if err := stmt.GetContext(ctx, &u.ID, u); err != nil {
		return nil, fmt.Errorf("error getting inserted ID: %w", err)
	}

	return u, nil

}

func (s *Storage) GetUser(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	query := `SELECT * FROM users WHERE email=$1`
	err := s.db.GetContext(ctx, &u, query, email)
	if err != nil {
		return nil, fmt.Errorf("error getting user data %w", err)
	}

	return &u, nil
}

func (s *Storage) ListUsers(ctx context.Context) ([]model.User, error) {
	var ul []model.User
	query := "SELECT * FROM users"
	err := s.db.GetContext(ctx, &ul, query)
	if err != nil {
		return nil, fmt.Errorf("error getting users data %w", err)
	}

	return ul, nil
}

func (s *Storage) UpdateUser(ctx context.Context, u *model.User) (*model.User, error) {
	query := "UPDATE users SET name=:name, email=:email, password=:password, is_admin=:is_admin, updated_at=:updated_at WHERE id=:id"
	_, err := s.db.NamedExecContext(ctx, query, u)
	if err != nil {
		return nil, fmt.Errorf("error updating user %w", err)
	}
	return u, nil
}

func (s *Storage) DeleteUser(ctx context.Context, id int64) error {
	query := "DELETE FROM users WHERE email=?"
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("error deleting user data %w", err)
	}

	return nil
}

// func (s *Storage) CreateUser(ctx context.Context, u *model.User) (*model.User, error) {
// 	query := "INSERT INTO users (name, email, password, is_admin) VALUES (:name, :email, :password, :is_admin)"
// 	res, err := s.db.NamedExecContext(ctx, query, u)
// 	if err != nil {
// 		return nil, fmt.Errorf("Error insering user %w", err)
// 	}
//
// 	id, err := res.LastInsertId()
// 	if err != nil {
// 		return nil, fmt.Errorf("Error getting last inserted id %w", err)
// 	}
// 	u.ID = id
//
// 	return u, nil
// }
