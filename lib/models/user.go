package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID             uint           `gorm:"primaryKey" json:"id"`
	Email          string         `gorm:"unique;not null" json:"email"`
	Password       string         `gorm:"not null" json:"password"`
	FullName       string         `gorm:"not null" json:"fullName"`
	BirthDay       time.Time      `gorm:"not null" json:"birthday"`
	Gender         string         `gorm:"size:1;not null" json:"gender"`
	PicURL         string         `gorm:"not null" json:"picURL"`
	EmailVerified  bool           `json:"emailVerified"`
	PhoneNumber    string         `json:"phoneNumber"`
	Nationality    string         `json:"nationality"`
	Address        string         `json:"address"`
	PostalCode     uint           `json:"postalCode"`
	IsAdmin        bool           `gorm:"not null" json:"isAdmin"`
	FidelityPoints uint           `gorm:"not null" json:"fidelityPoints"`
	AuthProviders  []AuthProvider `gorm:"many2many:user_auth_providers" json:"-"`
	Reservations   []Reservation  `gorm:"foreignKey:UserID" json:"reservations,omitempty"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

type AuthProvider struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Provider  string         `gorm:"not null" json:"provider"`
	Users     []User         `gorm:"many2many:user_auth_providers" json:"users,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

func (user User) ValidateRegistration() error {
	if user.Email == "" {
		return errors.New("EMAIL_UNDEFINED")
	}
	if user.Password == "" {
		return errors.New("PASSWORD_UNDEFINED")
	}
	if user.FullName == "" {
		return errors.New("FULLNAME_UNDEFINED")
	}
	if user.Gender != "F" && user.Gender != "M" {
		return errors.New("GENDER_UNDEFINED")
	}
	if user.PicURL == "" {
		return errors.New("PICURL_UNDEFINED")
	}
	if !isAgeValid(user.BirthDay) {
		return errors.New("BIRTHDAY_NOT_ALLOWED")
	}
	return nil
}

func isAgeValid(birthDay time.Time) bool {
	isAgeSmallerThanFive := birthDay.After(time.Now().AddDate(-5, 0, 0))
	isAgeGreaterThanHundred := birthDay.Before(time.Now().AddDate(-100, 0, 0))

	isAgeNotValide := isAgeSmallerThanFive || isAgeGreaterThanHundred
	return !isAgeNotValide
}

func (user User) ValidateLogin() error {
	if user.Email == "" {
		return errors.New("EMAIL_UNDEFINED")
	}
	if user.Password == "" {
		return errors.New("PASSWORD_UNDEFINED")
	}
	return nil
}
