package main

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"
	appdb "github.com/ppxmmm/miniBigCProject_Backend/internal/db"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/util"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables and defaults")
	}

	config := util.LoadConfig()
	database, err := gorm.Open(postgres.Open(config.DB.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("get database connection pool: %v", err)
	}
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatalf("connect database: %v", err)
	}

	if err := appdb.Bootstrap(ctx, database); err != nil {
		log.Fatalf("seed database: %v", err)
	}

	log.Println("Database seeded successfully")
}
