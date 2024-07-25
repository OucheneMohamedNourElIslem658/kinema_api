package reservations

import (
	"net/http"

	authMiddlewares "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/features/auth/middlewares"
	reservationsControllers "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/features/reservations/controllers"
	reservationsSockets "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/features/reservations/sockets"
	"github.com/OucheneMohamedNourElIslem658/kinema_api/lib/tools"
)

type ReservationsRouter struct {
	Router                  *http.ServeMux
	SeatChoiceSocketManager *reservationsSockets.SeatChoiceSocketManager
	authMiddlewares         *authMiddlewares.AuthMiddlewares
	ReservationsContoller   *reservationsControllers.ReservationsController
}

func NewReservationsRouter() *ReservationsRouter {
	return &ReservationsRouter{
		Router:                  http.NewServeMux(),
		SeatChoiceSocketManager: reservationsSockets.NewSeatChoiceSocketManager(),
		authMiddlewares:         authMiddlewares.NewAuthMiddlewares(),
		ReservationsContoller:   reservationsControllers.NewReservationsController(),
	}
}

func (reservationsRouter *ReservationsRouter) RegisterRouts() {
	router := reservationsRouter.Router
	reservationController := reservationsRouter.ReservationsContoller
	seatChoiceSocketManager := reservationsRouter.SeatChoiceSocketManager
	middlewares := reservationsRouter.authMiddlewares

	authorizationWithEmailVerification := tools.MiddlewareChain(
		middlewares.Authorization,
		middlewares.AuthorizationWithEmailVerification,
	)

	authorizationWithAdminCheck := tools.MiddlewareChain(
		middlewares.Authorization,
		middlewares.AuthorizationWithEmailVerification,
		middlewares.AuthorizationWithAdminCheck,
	)

	router.HandleFunc("/seatChoice", authorizationWithEmailVerification(http.HandlerFunc(seatChoiceSocketManager.ServeWS)))
	router.HandleFunc("GET /getPaymentKeys", authorizationWithEmailVerification(http.HandlerFunc(reservationController.GetPaymentKeys)))
	router.HandleFunc("POST /createPaymentIntent", authorizationWithEmailVerification(http.HandlerFunc(reservationController.CreatePaymentIntent)))
	router.HandleFunc("POST /addReservation", authorizationWithEmailVerification(http.HandlerFunc(reservationController.AddReservation)))
	router.HandleFunc("DELETE /cancelReservation", authorizationWithEmailVerification(http.HandlerFunc(reservationController.CancelReservation)))
	router.HandleFunc("PUT /updateReservation", authorizationWithAdminCheck(http.HandlerFunc(reservationController.UpdateReservation)))
	router.HandleFunc("POST /getReservations", authorizationWithAdminCheck(http.HandlerFunc(reservationController.GetReservations)))
	router.HandleFunc("GET /getUserReservations", authorizationWithEmailVerification(http.HandlerFunc(reservationController.GetUserReservations)))
	router.HandleFunc("POST /getReservation", authorizationWithEmailVerification(http.HandlerFunc(reservationController.GetReservation)))
}