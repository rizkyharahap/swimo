.PHONY: help swagger swagger-force clean build run dev swagger-quick check-changes

# -------------------------------------------------------------------
# 🧭 Default target
help:
	@echo "Available targets:"
	@echo "  swagger        - Generate Swagger JSON, merge only new/changed paths"
	@echo "  swagger-force  - Force regenerate Swagger JSON (overwrite)"
	@echo "  clean          - Clean generated files"
	@echo "  build          - Build application"
	@echo "  run            - Build and run application"
	@echo "  dev            - Dev workflow (smart swagger + build + run)"
	@echo "  swagger-quick  - Quick swagger refresh (merge mode)"
	@echo "  check-changes  - Check if annotations changed"
# -------------------------------------------------------------------

SWAG_OUT=./docs/swagger/tmp
FINAL_JSON=./docs/swagger/swagger.json

# -------------------------------------------------------------------
# 🧩 Smart Swagger generation (merge only new/changed paths)
swagger:
	@echo "🔍 Checking for annotation changes..."
	@mkdir -p $(SWAG_OUT)
	@swag init -g ./cmd/app/main.go -o $(SWAG_OUT) --parseDependency --outputTypes json > /dev/null 2>&1 || true
	@if [ -f $(SWAG_OUT)/swagger.json ]; then \
		echo "🧠 Running Go merge tool (swagger-merge.go)..."; \
		go run ./swagger-merge.go --old $(FINAL_JSON) --new $(SWAG_OUT)/swagger.json; \
	else \
		echo "⚠️  No temporary swagger.json generated, skipping merge"; \
	fi
	@rm -rf $(SWAG_OUT)

# -------------------------------------------------------------------
# 🔁 Force regenerate (overwrite existing swagger.json)
swagger-force:
	@echo "⚙️  Force generating Swagger JSON documentation (overwrite existing file)"
	@swag init -g ./cmd/app/main.go -o ./docs/swagger/ --parseDependency --outputTypes json
	@echo "✅ Swagger JSON regenerated at $(FINAL_JSON)"

# -------------------------------------------------------------------
# 🧹 Clean temporary files
clean:
	@echo "🧼 Cleaning generated files..."
	@rm -rf ./docs/swagger/tmp ./docs/swagger/swagger.json.backup
	@echo "Done."

# -------------------------------------------------------------------
# 🏗️ Build app
build:
	@echo "🛠️  Building..."
	@go build -o ./main ./cmd/app/main.go
	@echo "✅ Build complete."

# -------------------------------------------------------------------
# 🚀 Run app
run: build
	@echo "🚀 Running..."
	@./main

# -------------------------------------------------------------------
# 🔄 Dev workflow (smart swagger + build + run)
dev: swagger build run

# -------------------------------------------------------------------
# ⚡ Quick swagger regeneration (merge mode)
swagger-quick:
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
