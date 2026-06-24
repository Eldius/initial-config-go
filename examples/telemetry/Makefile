run:
	go run main.go

stop:
	docker compose down

start: stop
	COMPOSE_BAKE=true docker compose up -d --build
