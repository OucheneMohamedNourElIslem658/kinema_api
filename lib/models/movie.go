package models

import (
	"time"

	"gorm.io/gorm"
)

type Movie struct {
	ID           uint           `gorm:"primaryKey" json:"id,omitempty"`
	Title        string         `gorm:"not null" json:"title,omitempty"`
	Description  string         `json:"description,omitempty"`
	Type         []Type         `gorm:"many2many:movie_types;" json:"type,omitempty"`
	Language     string         `json:"language,omitempty"`
	Cast         []Actor        `gorm:"many2many:movie_actors;" json:"cast,omitempty"`
	Rate         float64        `json:"rate,omitempty"`
	TrailerURL   string         `json:"trailerURL,omitempty"`
	Duration     time.Duration  `json:"duration,omitempty"`
	VoteCount    uint           `json:"voteCount,omitempty"`
	TrailerViews uint           `json:"trailerViews,omitempty"`
	PicURL       string         `json:"picURL,omitempty"`
	PO           string         `json:"p_o,omitempty"`
	Diffusions   []Diffusion    `gorm:"foreignKey:MovieID" json:"diffusions,omitempty"`
	CreatedAt    time.Time      `json:"-"`
	UpdatedAt    time.Time      `json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

type Actor struct {
	ID     uint    `gorm:"primaryKey" json:"id"`
	Name   string  `gorm:"unique;not null" json:"name"`
	PicURL string  `gorm:"type:text" json:"picURL"`
	Movies []Movie `gorm:"many2many:movie_actors;" json:"movies,omitempty"`
}

type Type struct {
	ID     uint    `gorm:"primaryKey" json:"id"`
	Name   string  `gorm:"unique,not null" json:"name"`
	Movies []Movie `gorm:"many2many:movie_types;" json:"movies,omitempty"`
}
