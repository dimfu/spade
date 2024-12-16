run: stop up

up:
	docker compose up -d --build

stop:
	docker compose stop

db-shell:
	docker exec -it spade-mysql mysql -u root -p