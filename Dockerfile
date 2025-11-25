# syntax=docker/dockerfile:1

FROM golang:1.25.1 AS builder
WORKDIR /app

ARG SQLC_VERSION=v1.30.0

# Install sqlc for database code generation
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@${SQLC_VERSION}

COPY . .

# Generate database code
RUN sqlc generate

# Tidy dependencies
RUN go mod tidy 

# Build the operations service
ENV CGO_ENABLED=0
ARG TARGETARCH
RUN GOOS=linux GOARCH=${TARGETARCH:-amd64} go build -o /app/bin/operations-server ./cmd/server

FROM gcr.io/distroless/base-debian12:nonroot
WORKDIR /app

COPY --from=builder /app/bin/operations-server /usr/local/bin/operations-server
COPY --from=builder /app/db /app/db

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/operations-server"]
