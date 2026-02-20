SHELL := /bin/bash

.PHONY: setup-models up down run migrate

setup-models:
	ollama pull llama3:8b
	ollama pull gemma2:9b
	ollama pull mistral:7b
	ollama pull qwen2.5:7b
	ollama pull llama3.1:8b

up:
	docker docker compose up

down:
	docker compose down

migrate:
	docker compose run --rm migrate

run:
	go run ./cmd/bot

