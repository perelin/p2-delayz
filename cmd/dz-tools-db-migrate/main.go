package main

import (
	"fmt"
	"log"
	"os"
	models "p2lab/p2-delayz/pkg/models"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../../.env")
	if err != nil {
		fmt.Println("Error loading local .env file")
	}

	db, err := gorm.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Println(err)
		panic("failed to connect database")
	}
	defer db.Close()

	db.AutoMigrate(&models.DzStationSchedule{})
}
