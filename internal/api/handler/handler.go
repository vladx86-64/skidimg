package handler

import (
	"context"
	server "skidimg/internal/service"
	"skidimg/internal/token"
)

type handler struct {
	ctx        context.Context
	server     *server.Server
	TokenMaker *token.JWTMaker
}

func NewHandler(server *server.Server, secretKey string) *handler {
	return &handler{
		ctx:        context.Background(),
		server:     server,
		TokenMaker: token.NewJWTMaker(secretKey),
	}
}
