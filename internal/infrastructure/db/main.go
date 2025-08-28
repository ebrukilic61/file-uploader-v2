package db

import (
	"log"
)

func main() {
	database, err := NewPostgresDB()
	if err != nil {
		log.Fatal("DB connection failed:", err)
	} else {
		log.Println("DB connection established successfully!")
	}

	err = AutoMigrate(database)
	if err != nil {
		log.Fatal("AutoMigrate failed:", err)
	}

	log.Println("Database migrated successfully!")
}
