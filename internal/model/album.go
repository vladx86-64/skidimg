package model

import "time"

type Album struct {
	ID          int64      `db:"id" json:"id"`
	UserID      int64      `db:"user_id" json:"user_id"`
	Title       string     `db:"title" json:"title"`
	Description string     `db:"description" json:"description"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at" json:"updated_at"`
}

type ImageAlbum struct {
	ImageID int64 `db:"image_id" json:"image_id"`
	AlbumID int64 `db:"album_id" json:"album_id"`
}
