package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongo() (*mongo.Database, error) {
	uri := os.Getenv("DATABASE_URI_1")
	dbName := os.Getenv("DATABASE_NAME_1")
	if uri == "" {
		log.Println("Warning: DATABASE_URI_1 not set, using default mongodb://localhost:27017")
		uri = "mongodb://localhost:27017"
	}
	if dbName == "" {
		log.Println("Warning: DATABASE_NAME_1 not set, using default db_prestasi")
		dbName = "db_prestasi"
	}

	clientOptions := options.Client().ApplyURI(uri)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}

	log.Printf("Connected to MongoDB %s (db: %s)", uri, dbName)
	return client.Database(dbName), nil
}

func ConnectPostgres() (*sql.DB, error) {
	uri := os.Getenv("DATABASE_URI_2")
	if uri == "" {
		log.Println("Warning: DATABASE_URI_2 not set, using default postgresql://nouval@localhost:5432/db_validasi?sslmode=disable")
		uri = "postgresql://nouval@localhost:5432/db_validasi?sslmode=disable"
	}

	db, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	log.Printf("Connected to Postgres: %s", uri)
	return db, nil
}

//koneksi ke dua database
func ConnectDB() (*mongo.Database, *sql.DB, error) {
	mDB, err := ConnectMongo()
	if err != nil {
		return nil, nil, fmt.Errorf("connect mongo: %w", err)
	}

	pgDB, err := ConnectPostgres()
	if err != nil {
		_ = mDB.Client().Disconnect(context.Background())
		return nil, nil, fmt.Errorf("connect postgres: %w", err)
	}

	return mDB, pgDB, nil
}
