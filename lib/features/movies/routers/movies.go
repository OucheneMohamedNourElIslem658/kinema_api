package movies

import (
	"net/http"

	authMiddlewares "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/features/auth/middlewares"
	moviesControllers "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/features/movies/controllers"
	"github.com/OucheneMohamedNourElIslem658/kinema_api/lib/tools"
)

type MoviesRouter struct {
	Router      *http.ServeMux
	controller  moviesControllers.MoviesController
	middlewares authMiddlewares.AuthMiddlewares
}

func NewAuthRouter() *MoviesRouter {
	return &MoviesRouter{
		Router:      http.NewServeMux(),
		controller:  *moviesControllers.Newcontroller(),
		middlewares: *authMiddlewares.NewAuthMiddlewares(),
	}
}

func (moviesRouter *MoviesRouter) RegisterRouts() {
	router := moviesRouter.Router
	controller := moviesRouter.controller
	middlewares := moviesRouter.middlewares

	authorizationWithEmailVerification := tools.MiddlewareChain(
		middlewares.Authorization,
		middlewares.AuthorizationWithEmailVerification,
    )
	authorizationWithAdminCheck := tools.MiddlewareChain(
		middlewares.Authorization,
		middlewares.AuthorizationWithEmailVerification,
		middlewares.AuthorizationWithAdminCheck,
	)

	router.HandleFunc("GET /getMoviesFromTMDB", authorizationWithEmailVerification(http.HandlerFunc(controller.GetMoviesFromTMDB)))
	router.HandleFunc("GET /getMovieTrailersFromTMDB", authorizationWithEmailVerification(http.HandlerFunc(controller.GetMovieTrailersFromTMDB)))
	router.HandleFunc("POST /addMovie", authorizationWithEmailVerification(http.HandlerFunc(controller.AddMovie)))
	router.HandleFunc("GET /getMovie", authorizationWithEmailVerification(http.HandlerFunc(controller.GetMovie)))
	router.HandleFunc("GET /getMovies", authorizationWithAdminCheck(http.HandlerFunc(controller.GetMovies)))
	router.HandleFunc("PUT /updateMovie", authorizationWithAdminCheck(http.HandlerFunc(controller.UpdateMovie)))
	router.HandleFunc("POST /addHall", authorizationWithAdminCheck(http.HandlerFunc(controller.AddHall)))
	router.HandleFunc("POST /addDiffusion", authorizationWithAdminCheck(http.HandlerFunc(controller.AddDiffusion)))
	router.HandleFunc("GET /getHalls", authorizationWithAdminCheck(http.HandlerFunc(controller.GetHalls)))
	router.HandleFunc("GET /getAllWeeksUntilNextYear", authorizationWithAdminCheck(http.HandlerFunc(controller.GetAllWeeksUntilNextYear)))
	router.HandleFunc("DELETE /deleteMovie", authorizationWithAdminCheck(http.HandlerFunc(controller.DeleteMovie)))
	router.HandleFunc("POST /getDiffusionsForAdmin", authorizationWithEmailVerification(http.HandlerFunc(controller.GetDiffusionsForAdmin)))
	router.HandleFunc("DELETE /deleteDiffusion", authorizationWithAdminCheck(http.HandlerFunc(controller.DeleteDiffusion)))
	router.HandleFunc("GET /getTopDiffusion", authorizationWithEmailVerification(http.HandlerFunc(controller.GetTopDiffusion)))
	router.HandleFunc("POST /getDiffusionsByDay", authorizationWithEmailVerification(http.HandlerFunc(controller.GetDiffusionsByDay)))
	router.HandleFunc("GET /getMostPopularDiffusionsTrailers", authorizationWithEmailVerification(http.HandlerFunc(controller.GetMostPopularDiffusionsTrailers)))
	router.HandleFunc("GET /getDiffusionsForUsers", authorizationWithEmailVerification(http.HandlerFunc(controller.GetDiffusionsForUsers)))
	router.HandleFunc("GET /getMoviesDiffusions", authorizationWithEmailVerification(http.HandlerFunc(controller.GetMoviesDiffusions)))
}
