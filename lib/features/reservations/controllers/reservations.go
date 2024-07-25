package reservations

import (
	"encoding/json"
	"net/http"
	"time"

	reservationsRepo "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/features/reservations/repositories"
	"github.com/OucheneMohamedNourElIslem658/kinema_api/lib/models"
)

type ReservationsController struct {
	reservationsRepo *reservationsRepo.ReservationsRepo
}

func NewReservationsController() *ReservationsController {
	return &ReservationsController{
		reservationsRepo: reservationsRepo.NewReservationsRepo(),
	}
}

func (reservationsController *ReservationsController) GetPaymentKeys(w http.ResponseWriter, r *http.Request) {
	reservationsRepo := reservationsController.reservationsRepo
	status, result := reservationsRepo.GetPaymentKeys()

	secretKeyCookie := &http.Cookie{
		Name:     "secretKey",
		Value:    result["secretKey"],
		HttpOnly: true,
	}
	publishableKeyCookie := &http.Cookie{
		Name:     "publishableKey",
		Value:    result["publishableKey"],
		HttpOnly: true,
	}
	http.SetCookie(w, secretKeyCookie)
	http.SetCookie(w, publishableKeyCookie)
	w.WriteHeader(status)
}

func (reservationsController *ReservationsController) CreatePaymentIntent(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	json.NewDecoder(r.Body).Decode(&body)

	amount := body["amount"].(float64)
	currency := body["currency"].(string)

	reservationsRepo := reservationsController.reservationsRepo
	status, result := reservationsRepo.CreatePaymentIntent(int64(amount), currency)

	w.WriteHeader(status)
	response, _ := json.Marshal(&result)
	w.Write(response)
}

func (reservationsController *ReservationsController) AddReservation(w http.ResponseWriter, r *http.Request) {
	var body models.Reservation
	json.NewDecoder(r.Body).Decode(&body)

	auth, _ := r.Context().Value("auth").(map[string]any)
	id := uint(auth["id"].(float64))
	body.UserID = id

	reservationsRepo := reservationsController.reservationsRepo

	status, result := reservationsRepo.AddReservation(body)

	w.WriteHeader(status)
	response, _ := json.Marshal(&result)
	w.Write(response)
}

func (reservationsController *ReservationsController) CancelReservation(w http.ResponseWriter, r *http.Request) {
	var body models.Reservation
	json.NewDecoder(r.Body).Decode(&body)

	auth, _ := r.Context().Value("auth").(map[string]any)
	id := uint(auth["id"].(float64))
	body.UserID = id

	reservationsRepo := reservationsController.reservationsRepo

	status, result := reservationsRepo.CancelReservation(body)

	w.WriteHeader(status)
	response, _ := json.Marshal(&result)
	w.Write(response)
}

func (reservationsController *ReservationsController) UpdateReservation(w http.ResponseWriter, r *http.Request) {
	var body models.Reservation
	json.NewDecoder(r.Body).Decode(&body)

	reservationsRepo := reservationsController.reservationsRepo

	status, result := reservationsRepo.UpdateReservation(body)

	w.WriteHeader(status)
	response, _ := json.Marshal(&result)
	w.Write(response)
}

func (reservationsController *ReservationsController) GetReservations(w http.ResponseWriter, r *http.Request) {
	var body struct {
		HallName   string    `json:"hallName"`
		MovieTitle string    `json:"movieTitle"`
		ShowTime   time.Time `json:"showTime"`
		IsExpired  bool      `json:"isExpired"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	reservationsRepo := reservationsController.reservationsRepo

	status, result := reservationsRepo.GetReservations(
		body.HallName,
		body.MovieTitle,
		body.ShowTime,
		body.IsExpired,
	)

	w.WriteHeader(status)
	response, _ := json.Marshal(&result)
	w.Write(response)
}

func (reservationsController *ReservationsController) GetUserReservations(w http.ResponseWriter, r *http.Request) {
	auth, _ := r.Context().Value("auth").(map[string]any)
	userID := uint(auth["id"].(float64))

	reservationsRepo := reservationsController.reservationsRepo

	status, result := reservationsRepo.GetUserReservations(userID)

	w.WriteHeader(status)
	response, _ := json.Marshal(&result)
	w.Write(response)
}

func (reservationsController *ReservationsController) GetReservation(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	json.NewDecoder(r.Body).Decode(&body)
	reservationID := uint(body["reservationID"].(float64))

	auth, _ := r.Context().Value("auth").(map[string]any)
	userID := uint(auth["id"].(float64))

	reservationsRepo := reservationsController.reservationsRepo

	status, result := reservationsRepo.GetReservation(userID, reservationID)
	if status == http.StatusOK {
		diffusion := result["reservation"].(models.Reservation)
		w.WriteHeader(status)
		reponse, _ := json.MarshalIndent(diffusion, "", "\t")
		w.Write(reponse)
		return
	}

	w.WriteHeader(status)
	response, _ := json.Marshal(&result)
	w.Write(response)
}