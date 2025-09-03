package main

import (
	"log"
	"os"
	"skidimg/internal/api/handler"
	"skidimg/internal/platform/db"
	server "skidimg/internal/service"
	"skidimg/internal/storage"
)

const minSKsize = 32

func main() {
	secretKey := os.Getenv("JWT_SECRET_KEY")
	if secretKey == "" {
		log.Println("[!] JWT_SECRET_KEY is not set!")
		os.Exit(1)
	}
	os.MkdirAll("uploads/original", 0755)
	os.MkdirAll("uploads/optimized", 0755)

	db, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("[!] Error opening database %v", err)
	}
	defer db.Close()
	log.Printf("[+] Successfully connected to database")

	stor := storage.NewStorage(db.GetDB())
	serv := server.NewServer(stor)
	hdlr := handler.NewHandler(serv, *&secretKey)
	handler.RegisterRoutes(hdlr)
	handler.Start(":8080")
}
