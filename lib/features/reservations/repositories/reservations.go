package reservations

import (
	"errors"
	"net/http"
	"strings"
	"time"

	models "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/models"
	mysql "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/services/mysql"
	stripepayment "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/services/stripe_payment"
	stripe "github.com/stripe/stripe-go"
	paymentintent "github.com/stripe/stripe-go/paymentintent"
	refund "github.com/stripe/stripe-go/refund"
	gorm "gorm.io/gorm"
)

type ReservationsRepo struct {
	database *gorm.DB
	payment  stripepayment.Config
}

func NewReservationsRepo() *ReservationsRepo {
	stripe.Key = stripepayment.Instance.SecretKey
	return &ReservationsRepo{
		database: mysql.Instance,
		payment:  stripepayment.Instance,
	}
}

func (reservationsRepo *ReservationsRepo) GetSeats(diffuionID uint) (map[string]interface{}, error) {
	if diffuionID <= 0 {
		return nil, errors.New("INVALID_ID")
	}

	database := reservationsRepo.database

	var diffusion models.Diffusion
	err := database.Where("id = ?", diffuionID).Preload("SeatsStatus").First(&diffusion).Error
	if err != nil {
		return nil, errors.New("FETCHING_DIFFUSION_FAILED")
	}

	var seats []*models.Seat
	for _, seat := range diffusion.SeatsStatus {
		seats = append(seats, &seat)
	}

	return map[string]interface{}{
		"count":     len(seats),
		"seats":     seats,
		"seatPrice": diffusion.SeatPrice,
	}, nil
}

func (reservationsRepo *ReservationsRepo) ResetSeats(uid uint, diffuionID uint) error {
	if diffuionID <= 0 {
		return errors.New("INVALID_ID")
	}

	database := reservationsRepo.database

	err := database.Model(models.Seat{}).Where("diffusion_id = ? and status = ? and user_id = ?", diffuionID, "onhold", uid).Update("status", "availble").Error
	if err != nil {
		return errors.New("RESETING_SEATS_FAILED")
	}

	return nil
}

func (reservationsRepo *ReservationsRepo) AddReservation(reservation models.Reservation) (int, map[string]string) {
	if err := reservation.ValidateAdd(); err != nil {
		return http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		}
	}

	// Validate payment intent:
	payment, err := paymentintent.Get(reservation.PaymentIntent, nil)
	if err != nil {
		return http.StatusBadRequest, map[string]string{
			"error": "INVALID_PAYMENT_INTENT",
		}
	}

	if payment.Status != stripe.PaymentIntentStatusSucceeded {
		return http.StatusBadRequest, map[string]string{
			"error": "PAYMENT_HAS_NOT_BEEN_EFFECTED",
		}
	}

	database := reservationsRepo.database

	// Validate seats:
	var seats []models.Seat
	for _, seat := range reservation.Seats {
		err := database.Where("id = ?", seat.ID).First(&seat).Error
		if err != nil {
			return http.StatusBadRequest, map[string]string{
				"error": "SEAT_DOESNT_EXIST",
			}
		}

		if seat.Status == "reserved" {
			return http.StatusBadRequest, map[string]string{
				"error": "SEAT_ALREADY_RESERVED",
			}
		}

		seats = append(seats, seat)
	}

	// Add seats:
	for _, seat := range seats {
		seat.UserID = &reservation.UserID
		seat.DiffusionID = reservation.DiffusionID
		seat.Status = "reserved"
		err := database.Save(&seat).Error
		if err != nil {
			continue
		}
	}
	reservation.Seats = seats

	// Add bill detail:
	reservation.PaymentMethod = string(payment.PaymentMethod.Type)
	reservation.Amount = uint(payment.Amount)
	reservation.Currency = payment.Currency

	// Create reservation:
	err = database.Create(&reservation).Error
	if err != nil {
		return http.StatusInternalServerError, map[string]string{
			"error": "CREATING_RESERVATION_FAILED",
		}
	}

	return http.StatusOK, map[string]string{
		"error": "RESERVATION_ADDED",
	}
}

func (reservationsRepo *ReservationsRepo) CancelReservation(reservation models.Reservation) (int, map[string]string) {
	if err := reservation.ValidateCancel(); err != nil {
		return http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		}
	}

	database := reservationsRepo.database

	// Fetching reservation:
	err := database.Where("id = ? and user_id = ?", reservation.ID, reservation.UserID).Preload("Seats").First(&reservation).Error
	if err != nil {
		return http.StatusBadRequest, map[string]string{
			"error": "FETCHING_RESERVATION_FAILED",
		}
	}

	// Reteive money:
	paymentIntentID := getPaymentIntentID(reservation.PaymentIntent)
	_, err = refund.New(&stripe.RefundParams{
		PaymentIntent: &paymentIntentID,
	})

	if err != nil {
		return http.StatusBadRequest, map[string]string{
			"error":   "REFUND_FAILED",
			"message": err.Error(),
		}
	}

	for _, seat := range reservation.Seats {
		seat.Status = "availble"
		seat.ReservationID = nil
		database.Save(&seat)
	}

	//Delete reservation:
	err = database.Where("id = ?", reservation.ID).Unscoped().Delete(&reservation).Error
	if err != nil {
		return http.StatusBadRequest, map[string]string{
			"error": "DELETING_RESERVATION_FAILED",
		}
	}

	return http.StatusOK, map[string]string{
		"error": "RESERVATION_CANCELED",
	}
}

func getPaymentIntentID(paymentIntent string) string {
	parts := strings.Split(paymentIntent, "_")
	paymentIntentID := parts[0] + "_" + parts[1]
	return paymentIntentID
}

func (reservationsRepo *ReservationsRepo) CreatePaymentIntent(amount int64, currency string) (int, map[string]string) {
	if amount <= 0 || currency == "" {
		return http.StatusBadRequest, map[string]string{
			"error": "INVALID_ARGS",
		}
	}

	payment, err := paymentintent.New(&stripe.PaymentIntentParams{
		Amount:   stripe.Int64(7000),
		Currency: stripe.String("usd"),
	})

	if err != nil {
		return http.StatusInternalServerError, map[string]string{
			"error": "CREATING_PAYMENT_INTENT_FAILED",
		}
	}

	return http.StatusOK, map[string]string{
		"paymentIntent": payment.ClientSecret,
	}
}

func (reservationsRepo *ReservationsRepo) GetPaymentKeys() (int, map[string]string) {
	payment := reservationsRepo.payment
	return http.StatusOK, map[string]string{
		"secretKey":      payment.SecretKey,
		"publishableKey": payment.PublishableKey,
	}
}

func (reservationsRepo *ReservationsRepo) UpdateReservation(newReservation models.Reservation) (int, map[string]interface{}) {
	// Validating id:
	if newReservation.ID == 0 {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "INVALID_ID",
		}
	}

	var reservation models.Reservation
	database := reservationsRepo.database

	err := database.Where("id = ?", newReservation.ID).First(&reservation).Error
	if err != nil {
		return http.StatusBadRequest, map[string]interface{}{
			"error": "FOUNDING_RESERVATION_FAILED",
		}
	}

	// Updating movie
	if newReservation.UserID != 0 {
		err := database.Where("id = ?", newReservation.UserID).First(&models.User{}).Error
		if err != nil {
			return http.StatusBadRequest, map[string]interface{}{
				"error": "INVALID_USER_ID",
			}
		}
		reservation.UserID = newReservation.UserID
	}
	if newReservation.DiffusionID != 0 {
		err := database.Where("id = ?", newReservation.DiffusionID).First(&models.Diffusion{}).Error
		if err != nil {
			return http.StatusBadRequest, map[string]interface{}{
				"error": "INVALID_DIFFUSION_ID",
			}
		}
		reservation.DiffusionID = newReservation.DiffusionID
	}
	reservation.HasCome = newReservation.HasCome

	err = database.Save(&reservation).Error
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "UPDATING_RESERVATION_FAILED",
		}
	}

	return http.StatusOK, map[string]interface{}{
		"message": "RESERVATION_UPDATED",
	}
}

func (reservationsRepo *ReservationsRepo) GetReservations(hallName string, movieTitle string, showTime time.Time, isExpired bool) (int, map[string]interface{}) {
	database := reservationsRepo.database

	result := make(map[string]interface{})

	query := database.Model(&models.Reservation{}).
		Preload("Seats").
		Preload("Diffusion").
		Preload("Diffusion.Movie", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "title")
		}).
		Preload("Diffusion.Hall", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		}).
		Joins("JOIN diffusions ON reservations.diffusion_id = diffusions.id").
		Joins("JOIN halls ON diffusions.hall_id = halls.id").
		Joins("JOIN movies ON diffusions.movie_id = movies.id")
	if hallName != "" {
		query = query.Where("halls.name = ?", hallName)
	}
	if movieTitle != "" {
		query = query.Where("movies.title = ?", movieTitle)
	}
	if !showTime.IsZero() {
		query = query.Where("diffusions.show_time = ?", showTime)
	}
	if isExpired {
		query = query.Where("reservations.has_come = ?", false)
	}

	var reservations []models.Reservation
	err := query.Find(&reservations).Error
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "FETCHING_RESERVATIONS_FAILED",
		}
	}

	// Clean responce:
	for index := range reservations {
		reservation := &reservations[index]
		reservation.DiffusionID = 0
		reservation.Diffusion.MovieID = 0
		reservation.Diffusion.HallID = 0
		reservation.UserID = 0
		reservation.Diffusion.SeatPrice = 0
		reservation.PaymentIntent = ""
	}

	result["reservations"] = reservations
	result["count"] = len(reservations)
	return http.StatusOK, result
}

func (reservationsRepo *ReservationsRepo) GetUserReservations(userID uint) (int, map[string]interface{}) {
	database := reservationsRepo.database

	result := make(map[string]interface{})

	query := database.Model(&models.Reservation{}).
		Where("user_id = ?", userID).
		Preload("Diffusion.Movie", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "title", "rate", "pic_url")
		}).
		Preload("Diffusion.Movie.Type")

	var reservations []models.Reservation
	err := query.Find(&reservations).Error
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "FETCHING_RESERVATIONS_FAILED",
		}
	}

	// Clean responce:
	for index := range reservations {
		reservation := &reservations[index]
		reservation.DiffusionID = 0
		reservation.Diffusion.MovieID = 0
		reservation.Diffusion.HallID = 0
		reservation.UserID = 0
		reservation.Diffusion.SeatPrice = 0
		reservation.PaymentIntent = ""
		reservation.Diffusion.Hall = nil
		reservation.Amount = 0
		reservation.Currency = ""
	}

	result["reservations"] = reservations
	result["count"] = len(reservations)
	return http.StatusOK, result
}

func (reservationsRepo *ReservationsRepo) GetReservation(userID uint, reservationID uint) (int, map[string]interface{}) {
	database := reservationsRepo.database

	query := database.Model(&models.Reservation{}).
		Where("id = ? and user_id = ?", reservationID, userID).
		Preload("Seats", func(db *gorm.DB) *gorm.DB {
			return db.Select("seat_row", "seat_column")
		}).
		Preload("Diffusion.Movie", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "title", "rate", "pic_url")
		}).
		Preload("Diffusion.Movie.Type").
		Preload("Diffusion.Hall", func(db *gorm.DB) *gorm.DB {
			return db.Select("id", "name")
		})

	var reservation models.Reservation
	err := query.First(&reservation).Error
	if err != nil {
		return http.StatusInternalServerError, map[string]interface{}{
			"error": "FETCHING_RESERVATION_FAILED",
		}
	}

	// Clean responce:
	reservation.DiffusionID = 0
	reservation.Diffusion.MovieID = 0
	reservation.Diffusion.HallID = 0
	reservation.UserID = 0
	reservation.Diffusion.SeatPrice = 0
	reservation.PaymentIntent = ""
	reservation.Amount = 0
	reservation.Currency = ""

	return http.StatusOK, map[string]interface{}{
		"reservation": reservation,
	}
}
