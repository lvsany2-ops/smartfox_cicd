package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"experiment-service/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "experiment_db")
	dbUser := getEnv("DB_USER", "root")
	dbPassword := getEnv("DB_PASSWORD", "123456")
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPassword, dbHost, dbPort, dbName)
	var db *gorm.DB
	var err error
	for i := range 10 {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("%d/10: failed to connect database: %v", i, err)
		time.Sleep(5 * time.Second)
		if i == 10 {
			os.Exit(1)
		}
	}

	db.AutoMigrate(&models.Experiment{}, &models.Attachment{}, &models.Question{}, &models.TestCase{})
	DB = db
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
