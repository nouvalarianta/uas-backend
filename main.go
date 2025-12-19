package main

import (
	"context"
	"log"
	"os"
	"uas-backend/config"
	"uas-backend/database"
	_ "uas-backend/docs"
	"uas-backend/utils"

	"github.com/joho/godotenv"
)

// @title UAS Backend - Achievement Management API
// @version 1.0
// @description API Documentation for Achievement Management System with dual database (PostgreSQL + MongoDB)
// @contact.name API Support
// @contact.email support@example.com
// @host localhost:3000
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("env tidak ditemukan")
	}

	mDB, pgDB, err := database.ConnectDB()
	if err != nil {
		log.Fatalf("gagal konek databse %v", err)
	}
	defer mDB.Client().Disconnect(context.Background())
	defer pgDB.Close()

	// Start token blacklist cleanup routine
	utils.StartCleanupRoutine()
	log.Println("Token blacklist cleanup routine started")

	app := config.NewApp(mDB, pgDB)

	db1Name := os.Getenv("DATABASE_NAME_1")
	log.Printf("Aplikasi berjalan dengan DATABASE_NAME: %s", db1Name)

	db2Name := os.Getenv("DATABASE_NAME_2")
	log.Printf("Aplikasi berjalan dengan DATABASE_NAME: %s", db2Name)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	log.Printf("Server is starting on port :%s", port)
	log.Fatal(app.Listen(":" + port))
}
