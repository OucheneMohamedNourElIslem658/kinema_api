package auth

import (
	"errors"
	"fmt"
	"net/http"

	authUtils "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/features/auth/utils"
	"github.com/OucheneMohamedNourElIslem658/kinema_api/lib/models"
	mysql "github.com/OucheneMohamedNourElIslem658/kinema_api/lib/services/mysql"
	"gorm.io/gorm"
)

type AuthRepo struct {
	database *gorm.DB
}

func NewAuthRepository() *AuthRepo {
	return &AuthRepo{
		database: mysql.Instance,
	}
}

func (authRepo *AuthRepo) RegisterWithEmailAndPassword(user *models.User) (int, map[string]string) {
	// Validate inputs
	if err := user.ValidateRegistration(); err != nil {
		return http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		}
	}

	database := authRepo.database
	// Check if email is in use:
	var exist bool
	err := database.Model(&models.User{}).Select("count(*) > 0").Where("email = ?", user.Email).Find(&exist).Error
	if err != nil {
		return http.StatusInternalServerError, map[string]string{
			"error": "FINDING_USER_FAILED",
		}
	}
	if exist {
		return http.StatusBadRequest, map[string]string{
			"error": "EMAIL_ALREADY_IN_USE",
		}
	}

	// Get password hash and email:
	user.Password, err = authUtils.HashPassword(user.Password)
	if err != nil {
		return http.StatusInternalServerError, map[string]string{
			"error": "PASSWORD_HASH_FAILED",
		}
	}

	// Adding auth provider
	err = addAuthProvider(user, "password", database)
	if err != nil {
		return http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		}
	}

	// Create User:
	err = database.Create(&user).Error
	if err != nil {
		return http.StatusInternalServerError, map[string]string{
			"error": "USER_CREATION_FAILED",
		}
	}
	return http.StatusOK, map[string]string{
		"message": "USER_CREATED",
	}
}

func addAuthProvider(user *models.User, providerName string, database *gorm.DB) error {
	var authProvider models.AuthProvider
	if err := database.Where("provider = ?", providerName).First(&authProvider).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			authProvider = models.AuthProvider{
				Provider: providerName,
			}
			if err := database.Create(&authProvider).Error; err != nil {
				return errors.New("ADDING_AUTH_PROVIDER_FAILED")
			}
		}
	}
	user.AuthProviders = append(user.AuthProviders, authProvider)
	return nil
}

func (authRepo *AuthRepo) LoginWithEmailAndPassword(user *models.User) (int, map[string]string) {
	// Validate inputs
	if err := user.ValidateLogin(); err != nil {
		return http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		}
	}
	database := authRepo.database
	password := user.Password
	email := user.Email

	// Check for email:
	var storedUser models.User
	err := database.Where("email = ?", email).First(&storedUser).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return http.StatusBadRequest, map[string]string{
				"error": "EMAIL_NOT_FOUND",
			}
		} else {
			return http.StatusInternalServerError, map[string]string{
				"error": "FINDING_USER_FAILED",
			}
		}
	}

	// Check password
	passwordMatches := authUtils.VerifyPasswordHash(password, storedUser.Password)
	if !passwordMatches {
		return http.StatusBadRequest, map[string]string{
			"error": "INCORRECT_PASSWORD",
		}
	}

	// Generating and sending idToken
	idToken, err := authUtils.CreateIdToken(
		storedUser.ID,
		storedUser.Email,
		storedUser.EmailVerified,
		storedUser.IsAdmin,
	)
	if err != nil {
		return http.StatusInternalServerError, map[string]string{
			"error": "GENERATING_IDTOKEN_FAILED",
		}
	}
	return http.StatusOK, map[string]string{
		"idToken": idToken,
	}
}

func (authRepo *AuthRepo) Authorization(authorization string) (int, map[string]any) {
	// Validate authorization:
	if authorization == "" {
		return http.StatusBadRequest, map[string]any{
			"error": "INDEFINED_AUTHORIZATION",
		}
	}

	// Validate idToken:
	idToken := authorization[len("Bearer "):]
	claims, err := authUtils.VerifyToken(idToken)
	if err != nil {
		return http.StatusUnauthorized, map[string]any{
			"error": "UNAUTHORIZED",
		}
	}

	return http.StatusOK, map[string]any{
		"email":         claims["email"],
		"id":            claims["id"],
		"emailVerified": claims["emailVerified"],
		"isAdmin":       claims["isAdmin"],
		"idToken":       idToken,
	}
}

func (authRepo *AuthRepo) AuthorizationWithEmailVerification(emailVerified bool) (int, map[string]any) {
	if !emailVerified {
		return http.StatusUnauthorized, map[string]any{
			"error": "UNAUTHORIZED",
		}
	}

	return http.StatusOK, nil
}

func (authRepo *AuthRepo) AuthorizationWithAdminCheck(isAdmin bool) (int, map[string]any) {
	if !isAdmin {
		return http.StatusUnauthorized, map[string]any{
			"error": "UNAUTHORIZED",
		}
	}

	return http.StatusOK, nil
}

func (authRepo *AuthRepo) GetUser(id uint) (int, map[string]any) {
	// Validate authorization:
	if id == 0 {
		return http.StatusBadRequest, map[string]any{
			"error": "INDEFINED_ID",
		}
	}

	database := authRepo.database

	// Getting user:
	var user models.User
	err := database.Where("id = ?", id).First(&user).Error
	if err != nil {
		return http.StatusInternalServerError, map[string]any{
			"error": "FINDING_USER_FAILED",
		}
	}

	return http.StatusOK, map[string]any{
		"user": user,
	}
}

func (authRepo *AuthRepo) SendEmailVerificationLink(toEmail string, url string) (int, map[string]any) {
	// Validate toEmail:
	if toEmail == "" {
		return http.StatusBadRequest, map[string]any{
			"error": "INDEFINED_ID",
		}
	}

	// generating id Token:
	idToken, err := authUtils.CreateIdToken(0, toEmail, false, false)
	if err != nil {
		return http.StatusInternalServerError, map[string]any{
			"error": "GENERATING_IDTOKEN_FAILED",
		}
	}

	// Sending email:
	verificationLink := url + "/" + idToken
	message := fmt.Sprintf("Subject: Email verification link!\nThis is email verification link from kinema\n%v\nif you do not have to do with it dont browse it!", verificationLink)

	err = authUtils.SendEmailMessage(toEmail, message)

	if err != nil {
		return http.StatusInternalServerError, map[string]any{
			"error": "SENDING_EMAIL_FAILED",
		}
	}

	return http.StatusOK, map[string]any{
		"message": "VERIFICATION_LINK_SENT",
	}
}

func (authRepo *AuthRepo) VerifyEmail(email string) (int, map[string]any) {
	// Validate id:
	if email == "" {
		return http.StatusBadRequest, map[string]any{
			"error": "INDEFINED_EMAIL",
		}
	}

	database := authRepo.database

	// Updating user:
	var user models.User
	err := database.Where("email = ?", email).First(&user).Error
	if err != nil {
		return http.StatusInternalServerError, map[string]any{
			"error": "FINDING_USER_FAILED",
		}
	}

	if user.EmailVerified {
		return http.StatusBadRequest, map[string]any{
			"error": "USER_ALREADY_VERIFIED",
		}
	}

	err = database.Model(&models.User{}).Where("email = ?", email).Update("email_verified", true).Error
	if err != nil {
		return http.StatusInternalServerError, map[string]any{
			"error": "UPDATING_USER_FAILED",
		}
	}

	return http.StatusOK, map[string]any{
		"message": "USER_VERIFIED",
	}
}

func (authRepo *AuthRepo) ResetPassword(email string, newPassword string) (int, map[string]any) {
	// Validate id:
	if email == "" {
		return http.StatusBadRequest, map[string]any{
			"error": "INDEFINED_EMAIL",
		}
	}

	database := authRepo.database

	// Hashing password:
	newPasswordHash, err := authUtils.HashPassword(newPassword)
	if err != nil {
		return http.StatusInternalServerError, map[string]any{
			"error": "PASSWORD_HASH_FAILED",
		}
	}

	// Updating user:
	err = database.Model(&models.User{}).Where("email = ?", email).Update("password", newPasswordHash).Error
	if err != nil {
		return http.StatusInternalServerError, map[string]any{
			"error": "UPDATING_USER_FAILED",
		}
	}

	return http.StatusOK, map[string]any{
		"message": "PASSWORD_CHANGED",
	}
}

func (authRepo *AuthRepo) SendPasswordResetLink(toEmail string, url string) (int, map[string]any) {
	// Validate toEmail:
	if toEmail == "" {
		return http.StatusBadRequest, map[string]any{
			"error": "INDEFINED_ID",
		}
	}

	// generating id Token:
	idToken, err := authUtils.CreateIdToken(0, toEmail, false, false)
	if err != nil {
		return http.StatusInternalServerError, map[string]any{
			"error": "GENERATING_IDTOKEN_FAILED",
		}
	}

	// Sending email:
	resetLink := url + "/" + idToken
	message := fmt.Sprintf("Subject: Password reset link!\nThis is password reset link from kinema\n%v\nif you do not have to do with it dont browse it!", resetLink)
	err = authUtils.SendEmailMessage(toEmail, message)

	if err != nil {
		return http.StatusInternalServerError, map[string]any{
			"error": "SENDING_EMAIL_FAILED",
		}
	}

	return http.StatusOK, map[string]any{
		"message": "RESET_PASSWORD_LINK_SENT",
	}
}
