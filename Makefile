ifneq (,$(wildcard ./.env))
    include .env
    export $(shell sed 's/=.*//' .env)
endif

run: stop up

up:
	docker compose up -d --build

stop:
	docker compose stop

db-shell:
	docker exec -it db mysql -u root -p spade

seed:
	@if [ -z "$(file)" ]; then \
		echo "Usage: make seed file=<filename>"; \
	else \
		cat ./database/seeds/$(file) | docker exec -i db mysql -u root -p$(MYSQL_ROOT_PASSWORD) spade; \
	fi

.PHONY: run, up, stop, db-shell