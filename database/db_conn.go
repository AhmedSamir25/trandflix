package database

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"trendflix/models"
)

var DbConn *gorm.DB

func ConnDB() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found, using environment variables")
	}

	dsn := strings.TrimSpace(os.Getenv("DB_DSN"))
	if dsn == "" {
		dbName := strings.TrimSpace(os.Getenv("DB_NAME"))
		dbUser := strings.TrimSpace(os.Getenv("DB_USER"))
		dbPassword := os.Getenv("DB_PASSWORD")
		dbHost := strings.TrimSpace(os.Getenv("DB_HOST"))
		dbPort := strings.TrimSpace(os.Getenv("DB_PORT"))

		if dbName == "" || dbUser == "" || dbHost == "" || dbPort == "" {
			panic("database configuration is incomplete")
		}

		dsn = fmt.Sprintf(
			"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			dbUser,
			dbPassword,
			dbHost,
			dbPort,
			dbName,
		)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Error),
	})
	if err != nil {
		panic(fmt.Sprintf("database connection failed: %v", err))
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(fmt.Sprintf("database handle failed: %v", err))
	}

	if err := sqlDB.Ping(); err != nil {
		panic(fmt.Sprintf("database ping failed: %v", err))
	}

	DbConn = db
}

func Migrate() {
	if DbConn == nil {
		panic("database is not connected")
	}

	err := DbConn.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Item{},
		&models.CategoryItem{},
		&models.ResetToken{},
		&models.Favorite{},
		&models.Review{},
		&models.Banner{},
		&models.WatchLater{},
		&models.UserList{},
		&models.UserListItem{},
	)
	if err != nil {
		panic(fmt.Sprintf("migration failed: %v", err))
	}
}
