package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Hall struct {
	ID           uint        `gorm:"primaryKey" json:"id,omitempty"`
	Diffusions   []Diffusion `gorm:"foreignKey:HallID" json:"diffussionID,omitempty"`
	Name         string      `gorm:"unique;not null" json:"name,omitempty"`
	RowsCount    uint        `json:"rowsCount,omitempty"`
	ColumnsCount uint        `json:"columnsCount,omitempty"`
}

type Diffusion struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	MovieID      uint           `gorm:"not null;constraint:OnDelete:CASCADE" json:"movieID,omitempty"`
	Movie        Movie          `gorm:"foreinKey:ID" json:"movie,omitempty"`
	ShowTime     time.Time      `gorm:"not null" json:"showTime,omitempty"`
	ShowDuration time.Duration  `gorm:"not null" json:"showDuration,omitempty"`
	HallID       uint           `gorm:"not null;constraint:OnDelete:CASCADE" json:"hallID,omitempty"`
	Hall         *Hall          `gorm:"foreinKey:ID" json:"hall,omitempty"`
	SeatPrice    float64        `gorm:"not null" json:"seatPrice,omitempty"`
	CreatedAt    time.Time      `json:"-"`
	UpdatedAt    time.Time      `json:"-"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	SeatsStatus  []Seat         `gorm:"foreignKey:DiffusionID" json:"status,omitempty"`
	Reservations []Reservation  `gorm:"foreignKey:DiffusionID" json:"reservations,omitempty"`
}

type Seat struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	UserID        *uint  `gorm:"constraint:OnDelete:SET NULL" json:"userID,omitempty"`
	DiffusionID   uint   `gorm:"not null;constraint:OnDelete:CASCADE" json:"diffusionId"`
	ReservationID *uint  `gorm:"constraint:OnDelete:SET NULL" json:"reservationID,omitempty"`
	Status        string `gorm:"not null" json:"status"`
	SeatRow       string `gorm:"size:1;not null" json:"row"`
	SeatColumn    int    `gorm:"not null" json:"column"`
}

func (diffusion *Diffusion) Validate() error {
	if diffusion.MovieID == 0 {
		return errors.New("INVALID_MOVIE_ID")
	}
	if diffusion.ShowTime.Before(time.Now()) {
		return errors.New("INVALID_SHOW_TIME")
	}
	if diffusion.ShowDuration <= 0 {
		return errors.New("INVALID_SHOW_DURATION")
	}
	if diffusion.SeatPrice <= 0 {
		return errors.New("INVALID_SEAT_PRICE")
	}
	if diffusion.HallID == 0 {
		return errors.New("INVALID_HALL_ID")
	}
	return nil
}

func (hall *Hall) ValidateHall() error {
	if hall.Name == "" {
		return errors.New("INVALID_HALL_NAME")
	}
	if hall.RowsCount == 0 {
		return errors.New("INVALID_ROWS_COUNT")
	}
	if hall.ColumnsCount == 0 {
		return errors.New("INVALID_COLUMNS_COUNT")
	}
	return nil
}

func (seat *Seat) ValidateSeat() error {
	if seat.ID == 0 {
		return errors.New("INVALID_SEAT_ID")
	}
	return nil
}
