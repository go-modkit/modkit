package app

import (
	"encoding/json"
	"net/http"

	modkithttp "github.com/go-modkit/modkit/modkit/http"
)

type Controller struct{}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) RegisterRoutes(router modkithttp.Router) {
	router.Handle(http.MethodGet, "/health", http.HandlerFunc(c.handleHealth))
}

// @Summary Health check
// @Description Returns service health status.
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (c *Controller) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
}
