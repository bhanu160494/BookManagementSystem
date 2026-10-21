package controller

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"strings"
	"unicode/utf8"

	"Go-Learning/auth"
	"Go-Learning/models"
)

const minPasswordLength = 8

// AuthController exposes signup/signin endpoints backed by Keycloak.
type AuthController struct {
	keycloak *auth.KeycloakClient
}

// NewAuthController wires the handlers to a Keycloak client.
func NewAuthController(keycloak *auth.KeycloakClient) *AuthController {
	return &AuthController{keycloak: keycloak}
}

// Signup registers a new user in Keycloak.
func (a *AuthController) Signup(w http.ResponseWriter, r *http.Request) {
	var req models.SignupRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if err := validateSignup(req); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID, err := a.keycloak.CreateUser(r.Context(), auth.NewUser{
		Username:  req.Username,
		Email:     req.Email,
		FirstName: strings.TrimSpace(req.FirstName),
		LastName:  strings.TrimSpace(req.LastName),
		Password:  req.Password,
	})
	switch {
	case errors.Is(err, auth.ErrUserExists):
		writeJSONError(w, http.StatusConflict, "a user with that username or email already exists")
		return
	case err != nil:
		log.Println("signup failed:", err)
		writeJSONError(w, http.StatusBadGateway, "unable to register user at this time")
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"id":       userID,
		"username": req.Username,
		"message":  "user registered successfully",
	})
}

// Signin exchanges credentials for a Keycloak access token.
func (a *AuthController) Signin(w http.ResponseWriter, r *http.Request) {
	var req models.SigninRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	token, err := a.keycloak.Login(r.Context(), req.Username, req.Password)
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeJSONError(w, http.StatusUnauthorized, "invalid username or password")
		return
	case err != nil:
		log.Println("signin failed:", err)
		writeJSONError(w, http.StatusBadGateway, "unable to sign in at this time")
		return
	}

	writeJSON(w, http.StatusOK, token)
}

// Refresh issues a new access token from a refresh token.
func (a *AuthController) Refresh(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		return
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		writeJSONError(w, http.StatusBadRequest, "refreshToken is required")
		return
	}

	token, err := a.keycloak.Refresh(r.Context(), req.RefreshToken)
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeJSONError(w, http.StatusUnauthorized, "refresh token is invalid or expired")
		return
	case err != nil:
		log.Println("token refresh failed:", err)
		writeJSONError(w, http.StatusBadGateway, "unable to refresh token at this time")
		return
	}

	writeJSON(w, http.StatusOK, token)
}

// Logout revokes the supplied refresh token.
func (a *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		return
	}
	if strings.TrimSpace(req.RefreshToken) == "" {
		writeJSONError(w, http.StatusBadRequest, "refreshToken is required")
		return
	}

	if err := a.keycloak.Logout(r.Context(), req.RefreshToken); err != nil {
		log.Println("logout failed:", err)
		writeJSONError(w, http.StatusBadGateway, "unable to log out at this time")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// Me returns the profile derived from the caller's access token.
func (a *AuthController) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeJSONError(w, http.StatusUnauthorized, "missing authentication context")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"id":       claims.Subject,
		"username": claims.Username,
		"email":    claims.Email,
		"roles":    claims.Roles,
	})
}

func validateSignup(req models.SignupRequest) error {
	if req.Username == "" {
		return errors.New("username is required")
	}
	if utf8.RuneCountInString(req.Username) > 100 {
		return errors.New("username must be at most 100 characters")
	}
	if req.Email == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return errors.New("email is not a valid address")
	}
	if utf8.RuneCountInString(req.Password) < minPasswordLength {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}

func decodeJSONBody(w http.ResponseWriter, r *http.Request, target any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeJSONError(w, http.StatusBadRequest, "request body is not valid JSON")
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
