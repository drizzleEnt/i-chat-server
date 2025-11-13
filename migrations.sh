#!/bin/bash
source .env

export MIGRATION_DSN="host=postgres-chat-container port=$PG_PORT dbname=$POSTGRES_DB user=$POSTGRES_USER password=$POSTGRES_PASSWORD sslmode=$PG_SSL" #for local up
# export MIGRATION_DSN="host=$PG_HOST port=$PG_PORT dbname=$POSTGRES_DB user=$POSTGRES_USER password=$POSTGRES_PASSWORD sslmode=$PG_SSL"

sleep 2 && goose -dir "${MIGRATION_DIR}" postgres "${MIGRATION_DSN}" up -v