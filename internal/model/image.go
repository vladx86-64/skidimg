package model

import "time"

type ImageReq struct {
	UserID   int64  `json:"user_id"`
	Filename string `json:"filename"`
	Ext      string `json:"orig_ext"`
}

type ImageRes struct {
	ID       int64  `json:"id"`
	Filename string `json:"filename"`
}

type Image struct {
	ID        int64     `db:"id" json:"id"`
	UserID    *int64    `db:"user_id"`
	Filename  string    `db:"filename" json:"filename"`
	Ext       string    `db:"orig_ext" json:"orig_ext"`
	CreatedAt time.Time `db:"created_at"`
}

// func (irq *ImageReq) ToStorage() *Image {
// 	return &Image{
// 		UserID:   irq.UserID,
// 		Filename: irq.Filename,
// 		Ext:      irq.Ext,
// 	}
// }
