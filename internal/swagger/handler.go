package swagger

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/rizkyharahap/swimo/config"
	"github.com/rizkyharahap/swimo/pkg/logger"
	httpSwagger "github.com/swaggo/http-swagger"
)

type SwaggerHandler struct {
	cfg     *config.Config
	log     *logger.Logger
	Handler http.HandlerFunc
}

func NewSwaggerHandler(cfg *config.Config, log *logger.Logger) *SwaggerHandler {
	Handler := httpSwagger.Handler(
		httpSwagger.URL("/swagger/docs"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("list"),
		httpSwagger.DomID("swagger-ui"),
	)

	return &SwaggerHandler{cfg, log, Handler}
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

	// Update host and schemes if HTTP_BASE_URL is provided
	urlParts := strings.SplitN(h.cfg.HTTP.BaseURL, "://", 2)

	schemes := []string{urlParts[0]}
	host := urlParts[1]

	swaggerJSON["host"] = host
	swaggerJSON["schemes"] = schemes
	h.log.Info("Swagger host dynamically configured", "host", host, "schemes", schemes)

	// Marshal back to JSON
	updatedData, err := json.MarshalIndent(swaggerJSON, "", "    ")
	if err != nil {
		http.Error(w, "Failed to generate swagger.json", http.StatusInternalServerError)
		return
	}

	// Write the response
	w.Write(updatedData)
}
