run: stop up

up:
	docker compose up -d --build

stop:
	docker compose stop

db-shell:
	docker exec -it db mysql -u root -p

.PHONY: run, up, stop, db-shell