package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type DzStop struct {
	Station       string
	Train         string
	Direction     string
	TimeDeparture time.Time
	Platform      string
	TrainURL      string
	SourceURL     string
	gorm.Model
}
