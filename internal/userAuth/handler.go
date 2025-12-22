package userauth

import (
	"net/http"

	"github.com/varnit-ta/PlacementLog/pkg/utils"
)

/*
UserAuthHandler handles user authentication HTTP requests.
Provides endpoints for user login, registration, and logout.
*/
type UserAuthHandler struct {
	srv *UserAuthService
}

/*
requestPayload represents the JSON payload for login and registration requests.
*/
type loginRequestPayload struct {
	Regno    string `json:"regno"`
	Password string `json:"password"`
}

type registerRequestPayload struct {
	Regno    string `json:"regno"`
	Username string `json:"username"`
	Password string `json:"password"`
}

/*
responsePayload represents the JSON response for successful authentication.
*/
type responsePayload struct {
	UserID   string `json:"userid"`
	Regno    string `json:"regno"`
	Username string `json:"username"`
	Token    string `json:"token"`
}

/*
NewUserAuthHandler creates a new UserAuthHandler instance with the provided service.

Parameters:
- srv: The user authentication service

Returns:
- *UserAuthHandler: A new handler instance
*/
func NewUserAuthHandler(srv *UserAuthService) *UserAuthHandler {
	return &UserAuthHandler{
		srv: srv,
	}
}

/*
Login handles user login requests.
Now expects regno and password.

HTTP Method: POST
Endpoint: /auth/login

Request Body:

	{
	  "regno": "22bcs1234",
	  "password": "password123"
	}

Response (200 OK):

	{
	  "userid": "user_id",
	  "regno": "22bcs1234",
	  "name": "",
	  "token": "jwt_token_here"
	}

Returns:
- 200 OK: Successful login with token
- 400 Bad Request: Invalid request format
- 401 Unauthorized: Invalid credentials
*/
func (h *UserAuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var payload loginRequestPayload

	if err := utils.ReadJSON(r, &payload); err != nil {
		utils.WriteError(w, err)
		return
	}

	token, user, err := h.srv.Login(payload.Regno, payload.Password)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	resp := responsePayload{
		UserID:   user.ID,
		Regno:    user.Regno,
		Username: user.Username,
		Token:    token,
	}

	utils.WriteJSON(w, resp, http.StatusOK)
}

/*
Register handles user registration requests.
Now expects regno, name, and password.

HTTP Method: POST
Endpoint: /auth/register

Request Body:

	{
	  "regno": "22bcs1234",
	  "name": "John Doe",
	  "password": "password123"
	}

Response (201 Created):

	{
	  "userid": "user_id",
	  "regno": "22bcs1234",
	  "name": "John Doe",
	  "token": "jwt_token_here"
	}

Returns:
- 201 Created: Successful registration with token
- 400 Bad Request: Invalid request format
- 409 Conflict: Username already exists
*/
func (h *UserAuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var payload registerRequestPayload

	if err := utils.ReadJSON(r, &payload); err != nil {
		utils.WriteError(w, err)
		return
	}

	token, userId, err := h.srv.Register(payload.Regno, payload.Username, payload.Password)
	if err != nil {
		utils.WriteError(w, err)
		return
	}

	resp := responsePayload{
		UserID:   userId,
		Regno:    payload.Regno,
		Username: payload.Username,
		Token:    token,
	}

	utils.WriteJSON(w, resp, http.StatusCreated)
}
