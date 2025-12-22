package adminauth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/varnit-ta/PlacementLog/pkg/jwt"
	"github.com/varnit-ta/PlacementLog/pkg/utils"
)

/*
AdminAuthHandler handles admin authentication HTTP requests.
Provides endpoints for admin login, registration, and logout.
*/
type AdminAuthHandler struct {
	service *AdminService
}

/*
NewAdminAuthHandler creates a new AdminAuthHandler instance with the provided service.

Parameters:
- service: The admin authentication service

Returns:
- *AdminAuthHandler: A new handler instance
*/
func NewAdminAuthHandler(service *AdminService) *AdminAuthHandler {
	return &AdminAuthHandler{service: service}
}

/*
loginRequest represents the JSON payload for admin login and registration requests.
*/
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

/*
responsePayload represents the JSON response for successful admin authentication.
Returns both userid and username for better admin identification.
*/
type responsePayload struct {
	UserID   string `json:"userid"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

/*
Login handles admin login requests.
Validates admin credentials and returns a JWT token upon successful authentication.

HTTP Method: POST
Endpoint: /admin/login

Request Body:

	{
	  "username": "admin_username",
	  "password": "admin_password"
	}

Response (200 OK):

	{
	  "userid": "admin_uuid",
	  "username": "admin_username",
	  "token": "jwt_token_here"
	}

Returns:
- 200 OK: Successful login with token
- 400 Bad Request: Invalid request format
- 401 Unauthorized: Invalid credentials
*/
func (h AdminAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	token, admin, err := h.service.Login(req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	resp := responsePayload{
		UserID:   admin.ID,
		Username: admin.Username,
		Token:    token,
	}

	utils.WriteJSON(w, resp, http.StatusOK)
}

/*
Register handles admin registration requests.
Requires admin authentication to register new admins.
Only existing admins can register new admin accounts.

HTTP Method: POST
Endpoint: /admin/register

Headers Required:
- Authorization: Bearer <admin_jwt_token>

Request Body:

	{
	  "username": "new_admin_username",
	  "password": "new_admin_password"
	}

Response (201 Created):

	{
	  "userid": "new_admin_uuid",
	  "username": "new_admin_username",
	  "token": "jwt_token_here"
	}

Returns:
- 201 Created: Successful registration with token
- 400 Bad Request: Invalid request format
- 401 Unauthorized: Missing or invalid admin token
- 409 Conflict: Username already exists
*/
func (h AdminAuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "unauthorized: missing or invalid Authorization header", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	_, err := jwt.ValidateAdminToken(token)
	if err != nil {
		http.Error(w, "unauthorized: "+err.Error(), http.StatusUnauthorized)
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	tokenStr, admin, err := h.service.Register(req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	resp := responsePayload{
		UserID:   admin.ID,
		Username: admin.Username,
		Token:    tokenStr,
	}

	utils.WriteJSON(w, resp, http.StatusCreated)
}
