package db

import (
	"os"
	"p2lab/delayz/pkg/models"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func init() {

	err := godotenv.Load("../../.env")
	if err != nil {
		log.Error("Error loading local .env file")
	}

	logLevel, _ := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	log.SetLevel(logLevel)

}

func initDB() *gorm.DB {
	db, err := gorm.Open("postgres", os.Getenv("DATABASE_URL"))
	//db.LogMode(true)
	if err != nil {
		log.Println("failed to connect database", err)
		panic("failed to connect database")
	}
	return db
}

func SaveStopToDB(dzStop models.DzStop) {
	db := initDB()
	defer db.Close()
	//success := db.NewRecord(dzStop)
	db.Create(&dzStop)
	//log.Debug("db write success: ", success)
}
