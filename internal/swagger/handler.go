package swagger

import (
	"net/http"
	"strings"

	"github.com/rizkyharahap/swimo/config"
	"github.com/rizkyharahap/swimo/docs/swagger"
	_ "github.com/rizkyharahap/swimo/docs/swagger"
	httpSwagger "github.com/swaggo/http-swagger"
)

type SwaggerHandler struct {
	cfg     *config.Config
	Handler http.Handler
}

func NewSwaggerHandler(cfg *config.Config) *SwaggerHandler {
	urlParts := strings.SplitN(cfg.HTTP.BaseURL, "://", 2)

	if len(urlParts) == 2 {
		swagger.SwaggerInfo.Host = urlParts[1]
		swagger.SwaggerInfo.Schemes = []string{urlParts[0]}
	} else {
		// Fallback to default values
		swagger.SwaggerInfo.Host = "localhost:8080"
		swagger.SwaggerInfo.Schemes = []string{"http"}
	}

	return &SwaggerHandler{cfg: cfg, Handler: httpSwagger.Handler()}
}
