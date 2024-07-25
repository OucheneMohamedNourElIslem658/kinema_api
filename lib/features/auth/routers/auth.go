package auth

import (
	"net/http"

	authController "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/features/auth/controllers"
	authMiddlewares "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/features/auth/middlewares"
	"github.com/OucheneMohamedNourElIslem658/kinema_api/lib/tools"
)

type AuthRouter struct {
	Router      *http.ServeMux
	controller  authController.AuthController
	middlewares authMiddlewares.AuthMiddlewares
}

func NewAuthRouter() *AuthRouter {
	return &AuthRouter{
		Router:      http.NewServeMux(),
		controller:  *authController.Newcontroller(),
		middlewares: *authMiddlewares.NewAuthMiddlewares(),
	}
}

func (authRouter *AuthRouter) RegisterRouts() {
	router := authRouter.Router
	controller := authRouter.controller
	middlewares := authRouter.middlewares

	authorizationWithEmailVerification := tools.MiddlewareChain(
		middlewares.Authorization,
		middlewares.AuthorizationWithEmailVerification,
	)
	authorizationWithAdminCheck := tools.MiddlewareChain(
		middlewares.Authorization,
		middlewares.AuthorizationWithEmailVerification,
		middlewares.AuthorizationWithAdminCheck,
	)

	router.HandleFunc("POST /registerWithEmailAndPassword", controller.RegisterWithEmailAndPassword)
	router.HandleFunc("POST /loginWithEmailAndPassword", controller.LoginWithEmailAndPassword)
	router.HandleFunc("GET /getUser", authorizationWithEmailVerification(http.HandlerFunc(controller.GetUser)))
	router.HandleFunc("GET /getAdmin", authorizationWithAdminCheck(http.HandlerFunc(controller.GetUser)))
	router.HandleFunc("POST /sendEmailVerificationLink", controller.SendEmailVerificationLink)
	router.HandleFunc("GET /verifyEmail/{idToken}", controller.VerifyEmail)
	router.HandleFunc("POST /sendPasswordResetLink", controller.SendPasswordResetLink)
	router.HandleFunc("GET /serveResetPasswordForm/{idToken}", controller.ServeResetPasswordForm)
	router.HandleFunc("POST /resetPassword/{idToken}", controller.ResetPassword)
}
