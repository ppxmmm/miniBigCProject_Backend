package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	appdb "github.com/ppxmmm/miniBigCProject_Backend/internal/db"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/router"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/util"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables and defaults")
	}

	config := util.LoadConfig()

	db, err := gorm.Open(postgres.Open(config.DB.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatalf("open database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("get database connection pool: %v", err)
	}
	defer sqlDB.Close()

	sqlDB.SetConnMaxLifetime(5 * time.Minute)
	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("connect database: %v", err)
	}

	bootstrapCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := appdb.Bootstrap(bootstrapCtx, db); err != nil {
		log.Fatalf("bootstrap database: %v", err)
	}

	handler := router.New(config.AppName, db, config)

	log.Println("Server starting on port", config.Port)
	if err := http.ListenAndServe(":"+config.Port, handler); err != nil {
		log.Fatalf("start server: %v", err)
	}
}
