.PHONY: up down enable-undo disable-undo clean reset

# Start Postgres
up:
	docker compose up -d
	@echo "Waiting for Postgres..."
	@docker compose exec postgres pg_isready -U drift -d drift_example > /dev/null 2>&1 || sleep 2
	@echo "Postgres is ready."

# Stop Postgres and remove data
down:
	docker compose down -v

# Enable undo migrations by copying .bak files to .sql
enable-undo:
	@for f in migrations/*.sql.bak; do \
		cp "$$f" "$${f%.bak}"; \
	done
	@echo "Undo migrations enabled."

# Disable undo migrations
disable-undo:
	@rm -f migrations/U*.sql
	@echo "Undo migrations disabled."

# Remove generated files
clean:
	rm -f drift.yaml snapshot.json
	rm -f migrations/U*.sql

# Full reset: stop DB, remove data, clean generated files, restart
reset: down clean up
	@echo "Reset complete. Run 'drift init --driver postgres' to start over."
