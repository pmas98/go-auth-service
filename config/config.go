package config

import (
	"context"
	"log"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
)

var (
	DB  *gorm.DB
	ctx = context.Background()
)

func init() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
}

func Connect() {
	var err error
	dsn := os.Getenv("DB_DSN")
	DB, err = gorm.Open("postgres", dsn)
	if err != nil {
		panic("failed to connect to database")
	}

}

func GetDB() *gorm.DB {
	return DB
}

func GetContext() context.Context {
	return ctx
}
