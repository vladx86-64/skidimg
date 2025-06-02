package server

import (
	"skidimg/internal/storage"
)

type Server struct {
	storage *storage.Storage
}

func NewServer(storage *storage.Storage) *Server {
	return &Server{
		storage: storage,
	}
}
