.PHONY: help swagger swagger-force clean build run dev swagger-quick check-changes

# -------------------------------------------------------------------
# 🧭 Default target
help:
	@echo "Available targets:"
	@echo "  swagger        - Generate Swagger JSON, restore old examples into new file"
	@echo "  dev            - Dev workflow (swagger + build + run)"
# -------------------------------------------------------------------

SWAG_OUT=./docs/swagger

# -------------------------------------------------------------------
# 🧩 Generate Swagger and restore examples
swagger:
	@echo "⚡ Generating Swagger JSON and restoring examples..."
	@mkdir -p $(SWAG_OUT)
	@swag init -g ./cmd/app/main.go -o $(SWAG_OUT) --parseDependency --outputTypes go > /dev/null 2>&1 || true
	@echo "✅ Swagger JSON updated and examples restored."

# -------------------------------------------------------------------
# 🔄 Dev workflow (swagger + build + run with .env)
dev: swagger
	@echo "Loading environment variables from .env..."
	@export $$(grep -v '^#' .env | xargs) && go run ./cmd/app/main.go
