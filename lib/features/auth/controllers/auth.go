package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	authRepo "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/features/auth/repositories"
	models "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/models"
)

type AuthController struct {
	authRepo *authRepo.AuthRepo
}

func Newcontroller() *AuthController {
	return &AuthController{
		authRepo: authRepo.NewAuthRepository(),
	}
}

func (authcontroller *AuthController) RegisterWithEmailAndPassword(w http.ResponseWriter, r *http.Request) {
	var user models.User
	json.NewDecoder(r.Body).Decode(&user)

	authRepo := authcontroller.authRepo
	status, result := authRepo.RegisterWithEmailAndPassword(&user)

	w.WriteHeader(status)
	reponse, _ := json.Marshal(result)
	w.Write(reponse)
}

func (authcontroller *AuthController) LoginWithEmailAndPassword(w http.ResponseWriter, r *http.Request) {
	var user models.User
	json.NewDecoder(r.Body).Decode(&user)

	authRepo := authcontroller.authRepo
	status, result := authRepo.LoginWithEmailAndPassword(&user)

	if status == http.StatusOK {
		idTokenCookie := &http.Cookie{
			Name:     "idToken",
			Value:    result["idToken"],
			HttpOnly: true,
		}
		http.SetCookie(w, idTokenCookie)
		w.WriteHeader(status)
		return
	}
	w.WriteHeader(status)
	reponse, _ := json.Marshal(result)
	w.Write(reponse)
}

func (authcontroller *AuthController) GetUser(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	json.NewDecoder(r.Body).Decode(&body)

	authRepo := authcontroller.authRepo

	auth, _ := r.Context().Value("auth").(map[string]any)
	id := uint(auth["id"].(float64))
	status, result := authRepo.GetUser(id)

	if status == http.StatusOK {
		user := result["user"].(models.User)
		w.WriteHeader(status)
		response, _ := json.MarshalIndent(&user, "", "\t")
		w.Write(response)
		return
	}

	w.WriteHeader(status)
	reponse, _ := json.Marshal(result)
	w.Write(reponse)
}

func (authcontroller *AuthController) SendEmailVerificationLink(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	json.NewDecoder(r.Body).Decode(&body)

	authRepo := authcontroller.authRepo

	email := body["email"].(string)
	hostURL := "http://" + r.Host + "/api/v1/auth/verifyEmail"
	status, result := authRepo.SendEmailVerificationLink(email, hostURL)

	w.WriteHeader(status)
	reponse, _ := json.Marshal(result)
	w.Write(reponse)
}

func (authcontroller *AuthController) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	json.NewDecoder(r.Body).Decode(&body)

	authRepo := authcontroller.authRepo

	idToken := r.PathValue("idToken")
	authorization := fmt.Sprintf("Bearer %v", idToken)
	status, result := authRepo.Authorization(authorization)

	if status == http.StatusOK {
		email := result["email"].(string)
		status, result = authRepo.VerifyEmail(email)
	}

	w.WriteHeader(status)
	reponse, _ := json.Marshal(result)
	w.Write(reponse)
}

func (authcontroller *AuthController) SendPasswordResetLink(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	json.NewDecoder(r.Body).Decode(&body)

	authRepo := authcontroller.authRepo

	email := body["email"].(string)
	hostURL := "http://" + r.Host + "/api/v1/auth/serveResetPasswordForm"
	status, result := authRepo.SendPasswordResetLink(email, hostURL)

	w.WriteHeader(status)
	reponse, _ := json.Marshal(result)
	w.Write(reponse)
}

func (authcontroller *AuthController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	json.NewDecoder(r.Body).Decode(&body)

	authRepo := authcontroller.authRepo

	idToken := r.PathValue("idToken")
	authorization := fmt.Sprintf("Bearer %v", idToken)
	status, result := authRepo.Authorization(authorization)

	if status == http.StatusOK {
		email := result["email"].(string)
		newPassword := body["newPassword"].(string)
		status, result = authRepo.ResetPassword(email, newPassword)
	}

	w.WriteHeader(status)
	reponse, _ := json.Marshal(result)
	w.Write(reponse)
}

func (authcontroller *AuthController) ServeResetPasswordForm(w http.ResponseWriter, r *http.Request) {
	// Serve the HTML file
	formPath, err := filepath.Abs("./features/auth/views/reset_password_form.html")
	if err != nil {
		http.Error(w, "error finding html file", 500)
	}
	http.ServeFile(w, r, formPath)
}
