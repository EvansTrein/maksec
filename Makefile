GO=`which go`

.PHONY: build
build: build-agent build-service

.PHONY: build-agent
build-agent:
	cp ./config.yaml ./cmd/agent/config.yaml
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	$(GO) build -trimpath -ldflags="-s -w" -o ./bin/agent ./cmd/agent
	cp ./bin/agent ./internal/assets/bin/agent

.PHONY: build-service
build-service:
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags="-s -w" -o ./bin/maksec ./cmd/scripts

.PHONY: run
run: build-agent build-service
	./bin/maksec -c config.yaml

.PHONY: migrate-up
migrate-up:
	$(GO) run ./cmd/migrate -c config.yaml -u

.PHONY: stend-up
stend-up:
	docker compose up -d --build

.PHONY: stend-down
stend-down:
	docker compose down

.PHONY: stend-clean
stend-clean:
	docker compose down -v

.PHONY: test
test:
	$(GO) test -v -race -count=1 ./...

.PHONY: commit
commit: vet linter fmt

.PHONY: linter
linter:
	golangci-lint run ./...

.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: fmt
fmt:
	$(GO) fmt ./...
