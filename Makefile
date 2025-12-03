# Copyright (c) 2025 kk
#
# This software is released under the MIT License.
# https://opensource.org/licenses/MIT

.PHONY: help docker-build docker-up docker-down docker-logs docker-restart clean

help:
	@echo "Available commands:"
	@echo "  make docker-build    - Build Docker images"
	@echo "  make docker-up       - Start all services"
	@echo "  make docker-down     - Stop all services"
	@echo "  make docker-logs     - View logs"
	@echo "  make docker-restart  - Restart all services"
	@echo "  make clean           - Clean Docker resources"

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-restart:
	docker-compose restart

docker-ps:
	docker-compose ps

docker-clean:
	docker-compose down -v
	docker system prune -f

clean: docker-clean

