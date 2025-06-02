package model

import (
	"time"
)

type User struct {
	ID        int64      `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	Email     string     `db:"email" json:"email"`
	Password  string     `db:"password,omitempty" json:"-"`
	IsAdmin   bool       `db:"is_admin" json:"is_admin"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at"`
}

type UserReq struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	IsAdmin  bool   `json:"is_admin"`
}

type UserRes struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	IsAdmin bool   `json:"is_admin"`
}

// type ListUserReq struct {
//
// }

type ListUserRes struct {
	Users []UserRes `json:"users"`
}

type LoginUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserRes struct {
	SessionID             string    `json:"session_id"`
	AccessToken           string    `json:"access_token"`
	RefreshTOken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"acess_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"resfresh_token_expires_at"`
	User                  UserRes   `json:"user"`
}

type RenewAccessTokenReq struct {
	RefreshToken string `json:"refresh_token"`
}

type RenewAccessTokenRes struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"acess_token_expires_at"`
}

func (u *User) ToRes() *UserRes {
	return &UserRes{
		Name:    u.Name,
		Email:   u.Email,
		IsAdmin: u.IsAdmin,
	}
}

func (urq *UserReq) ToStorage() *User {
	return &User{
		Name:     urq.Name,
		Email:    urq.Email,
		Password: urq.Password,
		IsAdmin:  urq.IsAdmin,
	}
}
