package auth

import (
	"avito/internal/models"
	"avito/internal/storage"
	"avito/internal/utils"
	"avito/internal/utils/jwtToken"
	"context"
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

type Storage interface {
	GetUser(ctx context.Context, username string) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
}

type AuthController struct {
	Storage Storage
}

type AuthResponse struct {
	Token string `json:"token"`
}

type AuthRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (ac *AuthController) Auth(w http.ResponseWriter, r *http.Request) {
	var req AuthRequest

	validate := validator.New()
	err := json.NewDecoder(r.Body).Decode(&req)
	defer r.Body.Close()

	errValidate := validate.Struct(req)
	if err != nil || errValidate != nil {
		response := models.ErrorResponse{Errors: "the request parameters are incorrect"}
		utils.JsonResponse(w, http.StatusBadRequest, response)
		return
	}
	username := strings.ToLower(strings.TrimSpace(req.Username))
	const MinCost = 4
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), MinCost)
	if err != nil {
		response := models.ErrorResponse{Errors: "could not hash password"}
		utils.JsonResponse(w, http.StatusInternalServerError, response)
		return
	}

	user, err := ac.Storage.GetUser(r.Context(), username)
	if err != nil && !errors.Is(err, storage.ErrUserNotFound) {
		response := models.ErrorResponse{Errors: "internal server error"}
		utils.JsonResponse(w, http.StatusInternalServerError, response)
		return
	}

	if user == nil {
		newUser := &models.User{
			UUID:         uuid.New(),
			Balance:      1000,
			Username:     username,
			PasswordHash: string(hashedPassword),
			CreatedAt:    time.Now(),
		}

		err = ac.Storage.CreateUser(r.Context(), newUser)
		if err != nil {
			response := models.ErrorResponse{Errors: "could not create user"}
			utils.JsonResponse(w, http.StatusInternalServerError, response)
			return
		}

		SetCookie(w, newUser.UUID.String())
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		response := models.ErrorResponse{Errors: "wrong password"}
		utils.JsonResponse(w, http.StatusUnauthorized, response)
		return
	}

	SetCookie(w, user.UUID.String())
}

func SetCookie(w http.ResponseWriter, uuid string) {
	JwtToken, err := jwtToken.BuidToken(uuid)
	if err != nil {
		response := models.ErrorResponse{Errors: "could not create token"}
		utils.JsonResponse(w, http.StatusInternalServerError, response)
		return
	}

	cookie := &http.Cookie{
		Name:  "auth_token",
		Value: JwtToken,
	}
	http.SetCookie(w, cookie)
	response := AuthResponse{Token: JwtToken}
	utils.JsonResponse(w, http.StatusOK, response)
}
