FROM golang:1.26-alpine AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o internal/assets/bin/agent ./cmd/agent \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /bin/migrate ./cmd/migrate \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /bin/maksec ./cmd/scripts

FROM alpine:3.20
COPY --from=build /bin/maksec /usr/local/bin/maksec
COPY --from=build /bin/migrate /usr/local/bin/migrate

WORKDIR /app
COPY migrations ./migrations
COPY config.docker.yaml /etc/maksec/config.yaml

ENTRYPOINT ["/bin/sh", "-c", "/usr/local/bin/migrate -c /etc/maksec/config.yaml -u && exec /usr/local/bin/maksec -c /etc/maksec/config.yaml"]
