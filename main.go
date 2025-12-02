package main

import (
	"context"
	"uas-backend/config"
	"uas-backend/database"
	"log"
	"os"

	"github.com/joho/godotenv"
)

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
