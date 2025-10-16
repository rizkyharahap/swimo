package swagger

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/rizkyharahap/swimo/config"
	"github.com/rizkyharahap/swimo/pkg/logger"
	httpSwagger "github.com/swaggo/http-swagger"
)

type SwaggerHandler struct {
	cfg     *config.AppConfig
	logger  *logger.Logger
	Handler http.HandlerFunc
}

func NewSwaggerHandler(cfg *config.AppConfig, logger *logger.Logger) *SwaggerHandler {
	Handler := httpSwagger.Handler(
		httpSwagger.URL("/swagger/docs"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	)

	return &SwaggerHandler{cfg, logger, Handler}
}

func (h *SwaggerHandler) Docs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Read the swagger.json file
	swaggerData, err := os.ReadFile("docs/swagger/swagger.json")
	if err != nil {
		http.Error(w, "Failed to read swagger.json", http.StatusInternalServerError)
		return
	}

	// Parse JSON
	var swaggerJSON map[string]any
	if err := json.Unmarshal(swaggerData, &swaggerJSON); err != nil {
		http.Error(w, "Failed to parse swagger.json", http.StatusInternalServerError)
		return
	}

	// Update host if HTTP_BASE_URL is provided
	if h.cfg.Host != "" {
		swaggerJSON["host"] = h.cfg.Host
		h.logger.Info("Swagger host dynamically configured", "host", h.cfg.Host)
	}

	// Marshal back to JSON
	updatedData, err := json.MarshalIndent(swaggerJSON, "", "    ")
	if err != nil {
		http.Error(w, "Failed to generate swagger.json", http.StatusInternalServerError)
		return
	}

	// Write the response
	w.Write(updatedData)
}
