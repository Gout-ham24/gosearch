package api

import (
	"context"
	"encoding/json"
	"net/http"

	"gosearch/auth"
	"gosearch/storage"
)

type SignupRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	Email        string `json:"email"`
}

func (s *Server) SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "only POST allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || len(req.Password) < 8 {
		http.Error(w, `{"error": "email required, password must be at least 8 characters"}`, http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	existing, _ := storage.GetUserByEmail(ctx, s.Pool, req.Email)
	if existing != nil {
		http.Error(w, `{"error": "email already registered"}`, http.StatusConflict)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, `{"error": "failed to process password"}`, http.StatusInternalServerError)
		return
	}

	userID, err := storage.CreateUser(ctx, s.Pool, req.Email, hash)
	if err != nil {
		http.Error(w, `{"error": "failed to create user"}`, http.StatusInternalServerError)
		return
	}

	accessToken, _ := auth.GenerateAccessToken(userID, req.Email)
	refreshToken, _ := auth.GenerateRefreshToken()
	storage.StoreRefreshToken(ctx, s.Pool, userID, refreshToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Email:        req.Email,
	})
}

func (s *Server) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error": "only POST allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req SignupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	user, err := storage.GetUserByEmail(ctx, s.Pool, req.Email)
	if err != nil || user == nil || !auth.CheckPassword(req.Password, user.PasswordHash) {
		http.Error(w, `{"error": "invalid email or password"}`, http.StatusUnauthorized)
		return
	}

	accessToken, _ := auth.GenerateAccessToken(user.ID, user.Email)
	refreshToken, _ := auth.GenerateRefreshToken()
	storage.StoreRefreshToken(ctx, s.Pool, user.ID, refreshToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Email:        user.Email,
	})
}

// MeHandler handles GET /auth/me - returns the currently authenticated user's info.
func (s *Server) MeHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	user, err := storage.GetUserByID(context.Background(), s.Pool, userID)
	if err != nil || user == nil {
		http.Error(w, `{"error": "user not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	})
}

type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutHandler handles POST /auth/logout - invalidates one refresh token (this device only).
func (s *Server) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	storage.DeleteRefreshToken(context.Background(), s.Pool, req.RefreshToken)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out"})
}

// LogoutAllHandler handles POST /auth/logout-all - invalidates ALL sessions for the user.
func (s *Server) LogoutAllHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	storage.DeleteAllUserRefreshTokens(context.Background(), s.Pool, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "logged out from all devices"})
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ChangePasswordHandler handles PATCH /account/password
func (s *Server) ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error": "invalid request body"}`, http.StatusBadRequest)
		return
	}

	if len(req.NewPassword) < 8 {
		http.Error(w, `{"error": "new password must be at least 8 characters"}`, http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	user, err := storage.GetUserByID(ctx, s.Pool, userID)
	if err != nil || user == nil || !auth.CheckPassword(req.CurrentPassword, user.PasswordHash) {
		http.Error(w, `{"error": "current password is incorrect"}`, http.StatusUnauthorized)
		return
	}

	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, `{"error": "failed to process new password"}`, http.StatusInternalServerError)
		return
	}

	storage.UpdatePassword(ctx, s.Pool, userID, newHash)

	// Security best practice: invalidate all sessions after a password change
	storage.DeleteAllUserRefreshTokens(ctx, s.Pool, userID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "password changed, please log in again"})
}

// DeleteAccountHandler handles DELETE /account
func (s *Server) DeleteAccountHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, `{"error": "unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if err := storage.DeleteUser(context.Background(), s.Pool, userID); err != nil {
		http.Error(w, `{"error": "failed to delete account"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "account deleted"})
}
