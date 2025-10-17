package health

import (
	"fmt"
	"net/http"
	"time"

	"github.com/rizkyharahap/swimo/database"
	"github.com/rizkyharahap/swimo/pkg/logger"
)

type HealthHandler struct {
	log *logger.Logger
	db  *database.Database
}

func NewHealthHandler(log *logger.Logger, db *database.Database) *HealthHandler {
	return &HealthHandler{log, db}
}

// Health check
// @Summary      Health Check
// @Description  Check API and database connectivity
// @Tags         Monitoring
// @Produce json
// @Success 200  "Service healthy"
// @Failure 503  "Service unhealthy"
// @Router       /healthz [get]
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	w.Header().Set("Content-Type", "application/json")

	if h.db == nil {
		resp := fmt.Sprintf(`{"status":"unhealthy","timestamp":"%s","service":"swimo-api","database":"unconnected"}`,
			time.Now().UTC().Format(time.RFC3339))
		h.log.Error("Health check failed: database unconnected", resp)

		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	if err := h.db.Pool.Ping(ctx); err != nil {
		resp := fmt.Sprintf(`{"status":"unhealthy ping","timestamp":"%s","service":"swimo-api","database":"disconnected"}`,
			time.Now().UTC().Format(time.RFC3339))
		h.log.Error("Health check failed: ping error", resp)

		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	resp := fmt.Sprintf(`{"status":"healthy","timestamp":"%s","service":"swimo-api","database":"connected"}`,
		time.Now().UTC().Format(time.RFC3339))
	h.log.Info("Health check OK", "response", resp)

	w.WriteHeader(http.StatusOK)
}
