package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-modkit/modkit/examples/hello-mysql/internal/httpapi"
)

type Controller struct {
	service Service
}

func NewController(service Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) RegisterRoutes(router Router) {
	router.Handle(http.MethodGet, "/users/{id}", http.HandlerFunc(c.handleGetUser))
	router.Handle(http.MethodPost, "/users", http.HandlerFunc(c.handleCreateUser))
	router.Handle(http.MethodGet, "/users", http.HandlerFunc(c.handleListUsers))
	router.Handle(http.MethodPut, "/users/{id}", http.HandlerFunc(c.handleUpdateUser))
	router.Handle(http.MethodDelete, "/users/{id}", http.HandlerFunc(c.handleDeleteUser))
}

// @Summary Get user
// @Description Returns a user by id.
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} User
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Router /users/{id} [get]
func (c *Controller) handleGetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpapi.WriteProblem(w, r, http.StatusBadRequest, "invalid id")
		return
	}

	user, err := c.service.GetUser(r.Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteProblem(w, r, http.StatusNotFound, "not found")
			return
		}
		httpapi.WriteProblem(w, r, http.StatusInternalServerError, "internal error")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

// @Summary Create user
// @Description Creates a new user.
// @Tags users
// @Accept json
// @Produce json
// @Param body body CreateUserInput true "User payload"
// @Success 201 {object} User
// @Failure 400 {object} Problem
// @Failure 409 {object} Problem
// @Router /users [post]
func (c *Controller) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var input CreateUserInput
	if err := decodeJSON(r, &input); err != nil {
		httpapi.WriteProblem(w, r, http.StatusBadRequest, "invalid body")
		return
	}
	if input.Name == "" || input.Email == "" {
		httpapi.WriteProblem(w, r, http.StatusBadRequest, "name and email required")
		return
	}

	user, err := c.service.CreateUser(r.Context(), input)
	if err != nil {
		if errors.Is(err, ErrConflict) {
			httpapi.WriteProblem(w, r, http.StatusConflict, "user already exists")
			return
		}
		httpapi.WriteProblem(w, r, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

// @Summary List users
// @Description Returns all users.
// @Tags users
// @Produce json
// @Success 200 {array} User
// @Router /users [get]
func (c *Controller) handleListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := c.service.ListUsers(r.Context())
	if err != nil {
		httpapi.WriteProblem(w, r, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, users)
}

// @Summary Update user
// @Description Updates a user by id.
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param body body UpdateUserInput true "User payload"
// @Success 200 {object} User
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Router /users/{id} [put]
func (c *Controller) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpapi.WriteProblem(w, r, http.StatusBadRequest, "invalid id")
		return
	}

	var input UpdateUserInput
	if err := decodeJSON(r, &input); err != nil {
		httpapi.WriteProblem(w, r, http.StatusBadRequest, "invalid body")
		return
	}
	if input.Name == "" || input.Email == "" {
		httpapi.WriteProblem(w, r, http.StatusBadRequest, "name and email required")
		return
	}

	user, err := c.service.UpdateUser(r.Context(), id, input)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteProblem(w, r, http.StatusNotFound, "not found")
			return
		}
		httpapi.WriteProblem(w, r, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, user)
}

// @Summary Delete user
// @Description Deletes a user by id.
// @Tags users
// @Param id path int true "User ID"
// @Success 204 {object} map[string]string
// @Failure 400 {object} Problem
// @Failure 404 {object} Problem
// @Router /users/{id} [delete]
func (c *Controller) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		httpapi.WriteProblem(w, r, http.StatusBadRequest, "invalid id")
		return
	}

	if err := c.service.DeleteUser(r.Context(), id); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteProblem(w, r, http.StatusNotFound, "not found")
			return
		}
		httpapi.WriteProblem(w, r, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
