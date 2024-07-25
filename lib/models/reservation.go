package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Reservation struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	DiffusionID   uint           `gorm:"not null;constraint:OnDelete:CASCADE" json:"diffusionId,omitempty"`
	Diffusion     Diffusion      `gorm:"primaryKey:ID" json:"diffusion,omitempty"`
	UserID        uint           `gorm:"not null;constraint:OnDelete:CASCADE" json:"userId,omitempty"`
	Seats         []Seat         `gorm:"primaryKey:reservationID" json:"seats,omitempty"`
	HasCome       bool           `gorm:"not null" json:"hasCome"`
	PaymentMethod string         `gorm:"not null" json:"paymentMethod,omitempty"`
	Amount        uint           `gorm:"not null" json:"amount,omitempty"`
	Currency      string         `gorm:"not null" json:"currency,omitempty"`
	PaymentIntent string         `gorm:"unique;not null" json:"paymentIntent,omitempty"`
	CreatedAt     time.Time      `json:"-"`
	UpdatedAt     time.Time      `json:"-"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (reservation *Reservation) ValidateAdd() error {
	if reservation.UserID == 0 {
		return errors.New("INVALID_USER_ID")
	}
	if reservation.DiffusionID == 0 {
		return errors.New("INVALID_DIFFUSION_ID")
	}
	if reservation.Seats == nil {
		return errors.New("INVALID_SEATS")
	} else {
		for _, seat := range reservation.Seats {
			seat.ValidateSeat()
		}
	}
	return nil
}

func (reservation *Reservation) ValidateCancel() error {
	if reservation.ID == 0 {
		return errors.New("INVALID_ID")
	}
	if reservation.UserID == 0 {
		return errors.New("INVALID_USER_ID")
	}
	return nil
}
