.PHONY: build test vet check run seed test-api loadtest-smoke loadtest

build:
	go build ./...

test:
	go test ./... -count=1

vet:
	go vet ./...

# build + vet + unit tests (matches GitLab CI test job)
check: build vet test

run:
	go run ./cmd/server/main.go

seed:
	go run ./cmd/seed/main.go

# Playwright API tests (starts server + seeds DB if not already running)
test-api:
	chmod +x scripts/run-api-tests.sh
	./scripts/run-api-tests.sh

# k6 load tests (requires k6 or Docker)
loadtest-smoke:
	chmod +x scripts/run-loadtest.sh
	./scripts/run-loadtest.sh tests/load/smoke.js

loadtest:
	chmod +x scripts/run-loadtest.sh
	./scripts/run-loadtest.sh tests/load/dashboard.js
