.PHONY: help swagger swagger-force clean build run dev swagger-quick check-changes

# -------------------------------------------------------------------
# 🧭 Default target
help:
	@echo "Available targets:"
	@echo "  swagger        - Generate Swagger JSON, merge only new/changed paths"
	@echo "  swagger-force  - Force regenerate Swagger JSON (overwrite)"
	@echo "  dev            - Dev workflow (swagger + build + run)"
	@echo "  check-changes  - Check if annotations changed"
# -------------------------------------------------------------------

SWAG_OUT=./docs/swagger/tmp
FINAL_JSON=./docs/swagger/swagger.json

# -------------------------------------------------------------------
# 🧩 Smart Swagger generation (merge only new/changed paths)
swagger:
	@echo "⚡ Quick Swagger update (merge mode)..."
	@mkdir -p $(SWAG_OUT)
	@swag init -g ./cmd/app/main.go -o $(SWAG_OUT) --parseDependency --outputTypes json > /dev/null 2>&1 || true
	@if [ -f $(SWAG_OUT)/swagger.json ]; then \
		go run ./swagger-merge.go --old $(FINAL_JSON) --new $(SWAG_OUT)/swagger.json; \
	else \
		echo "⚠️  No temporary swagger.json generated"; \
	fi
	@rm -rf $(SWAG_OUT)

# -------------------------------------------------------------------
# 🔁 Force regenerate (overwrite existing swagger.json)
swagger-force:
	@echo "⚙️  Force generating Swagger JSON documentation (overwrite existing file)"
	@swag init -g ./cmd/app/main.go -o ./docs/swagger/ --parseDependency --outputTypes json
	@echo "✅ Swagger JSON regenerated at $(FINAL_JSON)"

# -------------------------------------------------------------------
# 🔄 Dev workflow (swagger + build + run with .env)
dev: swagger
	@echo "Loading environment variables from .env..."
	@export $$(grep -v '^#' .env | xargs) && go run ./cmd/app/main.go

# -------------------------------------------------------------------
# 🔍 Check if annotations have changed
check-changes:
	@echo "Checking for annotation changes..."
	@mkdir -p $(SWAG_OUT)
	@swag init -g ./cmd/app/main.go -o $(SWAG_OUT) --parseDependency --outputTypes json > /dev/null 2>&1 || true
	@if [ -f $(FINAL_JSON) ] && [ -f $(SWAG_OUT)/swagger.json ]; then \
		if diff $(FINAL_JSON) $(SWAG_OUT)/swagger.json > /dev/null 2>&1; then \
			echo "✅ No annotation changes detected"; \
		else \
			echo "⚠️  Annotation changes detected"; \
		fi; \
	elif [ -f $(SWAG_OUT)/swagger.json ]; then \
		echo "ℹ️  No existing swagger.json (first generation)"; \
	else \
		echo "❌ Failed to generate temporary swagger.json"; \
	fi
	@rm -rf $(SWAG_OUT)
