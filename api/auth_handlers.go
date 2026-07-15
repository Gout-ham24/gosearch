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
