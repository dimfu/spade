#!/bin/bash

if [[ -f .env ]]; then
    export $(grep -v '^#' .env | xargs)
else
    echo "Error: .env file not found."
    exit 1
fi

if [[ -z "$1" ]]; then
    echo "Error: Missing required argument."
    usage
fi

REQUIRED_VARS=("DB_USER" "DB_PASSWORD" "DB_HOST" "DB_PORT" "DB_NAME")
for VAR in "${REQUIRED_VARS[@]}"; do
    if [[ -z "${!VAR}" ]]; then
        echo "Error: $VAR is not set in the environment."
        exit 1
    fi
done

DIRECTION=$1
shift  
DOCKER_IMAGE="migrate/migrate"
MIGRATION_PATH="$(pwd)/database/migrations:/migrations"
NETWORK="spade_network"

case "$DIRECTION" in
    migrate-up)
        docker run --rm -v "$MIGRATION_PATH" --network "$NETWORK" "$DOCKER_IMAGE" \
            -path=/migrations \
            -database "mysql://$DB_USER:$DB_PASSWORD@tcp($DB_HOST:$DB_PORT)/$DB_NAME" \
            -verbose up "$@"
        ;;
    migrate-down)
        echo "y" | docker run --rm -v "$MIGRATION_PATH" --network "$NETWORK" "$DOCKER_IMAGE" \
            -path=/migrations \
            -database "mysql://$DB_USER:$DB_PASSWORD@tcp($DB_HOST:$DB_PORT)/$DB_NAME" \
            -verbose down -all
        ;;
    *)
        echo "Error: Invalid direction. Use 'migrate-up' or 'migrate-down'."
        usage
        ;;
esac

