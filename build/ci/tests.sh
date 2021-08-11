#!/bin/bash

export DOCKER_BUILDKIT=1
export BUILDKIT_PROGRESS=auto
export COMPOSE_DOCKER_CLI_BUILD=1

# Builds the application, runs the unit tests and starts the containers
docker-compose up -d --build

# Watching the logs
docker-compose logs -f &

# Waiting for the result of the e2e tests
docker-compose up --exit-code-from e2e_tests e2e_tests
STATUS=$?
docker-compose logs e2e_tests

# Stopping the containers (will also stop the logs watching process)
docker-compose down

exit $STATUS
