include .env
export

.PHONY: run
run:
	go run cmd/api/main.go