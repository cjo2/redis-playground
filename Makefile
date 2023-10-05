
.PHONY: redis-cli
redis-cli:
	@docker exec -it redis redis-cli

.PHONY: run-docker
run-docker:
	@docker-compose up

.PHONY: run-app
run-app:
	@go run ./cmd/main.go