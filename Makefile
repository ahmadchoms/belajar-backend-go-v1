run:
	docker compose up --build

stop:
	docker compose down

test:
	go test ./... -race -cover

clean:
	rm -f main
	docker compose down --volumes --remove-orphans
