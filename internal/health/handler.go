package health

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rizkyharahap/swimo/database"
	"github.com/rizkyharahap/swimo/pkg/logger"
	"github.com/rizkyharahap/swimo/pkg/response"
)

type HealthHandler struct {
	log *logger.Logger
	db  *database.Database
}

func NewHealthHandler(log *logger.Logger, db *database.Database) *HealthHandler {
	return &HealthHandler{log, db}
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if h.db == nil {
		resp := fmt.Sprintf(`{"status":"unhealthy","timestamp":"%s","service":"swimo-api","database":"unconnected"}`,
			time.Now().UTC().Format(time.RFC3339))
		h.log.Error("Health check failed: database unconnected", "response", resp)

		response.JSON(w, http.StatusServiceUnavailable, response.Message{Message: "Database unconnected"})
		return
	}

	if err := h.db.Pool.Ping(ctx); err != nil {
		resp := fmt.Sprintf(`{"status":"unhealthy ping","timestamp":"%s","service":"swimo-api","database":"disconnected"}`,
			time.Now().UTC().Format(time.RFC3339))
		h.log.Error("Health check failed: ping error", "response", resp)

		response.JSON(w, http.StatusServiceUnavailable, response.Message{Message: "Database ping failed"})
		return
	}

	resp := fmt.Sprintf(`{"status":"healthy","timestamp":"%s","service":"swimo-api","database":"connected"}`,
		time.Now().UTC().Format(time.RFC3339))
	h.log.Info("Health check OK", "response", resp)

	w.WriteHeader(http.StatusOK)
}
