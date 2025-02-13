# syntax=docker/dockerfile:1
FROM golang:1.23rc2-alpine3.20  AS build-stage
RUN apk add --no-progress --no-cache gcc musl-dev

WORKDIR /app

# Go modules
COPY go.mod go.sum ./
RUN go mod download

# Source code
COPY ./ ./

# Build
RUN go build -tags musl -ldflags '-extldflags "-static"' -o /server ./cmd

# Application binary into a lean image
FROM scratch AS build-release-stage

WORKDIR /app

COPY --from=build-stage /server /server

# Bind TCP port
EXPOSE 8090

# Run
ENTRYPOINT [ "/server" ]