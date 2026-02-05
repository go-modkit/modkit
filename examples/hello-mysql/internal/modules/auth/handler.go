package auth

import (
	"encoding/json"
	"net/http"

	"github.com/go-modkit/modkit/examples/hello-mysql/internal/httpapi"
)

type Handler struct {
	cfg Config
}

func NewHandler(cfg Config) *Handler {
	return &Handler{cfg: cfg}
}

func (h *Handler) RegisterRoutes(router Router) {
	router.Handle(http.MethodPost, "/auth/login", http.HandlerFunc(h.handleLogin))
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type tokenResponse struct {
	Token string `json:"token"`
}

func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var input loginRequest
	if err := decodeJSON(r, &input); err != nil {
		httpapi.WriteProblem(w, r, http.StatusBadRequest, "invalid body")
		return
	}

	if input.Username != h.cfg.Username || input.Password != h.cfg.Password {
		httpapi.WriteProblem(w, r, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token, err := IssueToken(h.cfg, User{ID: input.Username, Email: input.Username})
	if err != nil {
		httpapi.WriteProblem(w, r, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{Token: token})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func decodeJSON(r *http.Request, target any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
