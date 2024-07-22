# syntax=docker/dockerfile:1
FROM golang:1.21.4  AS build-stage

WORKDIR /app

# Go modules
COPY go.mod go.sum ./
RUN go mod download

# Source code
COPY ./ ./

# Build
RUN CGO_ENABLED=1 GOOS=linux go build -o /server ./cmd

# Run the tests in the container
FROM build-stage AS run-test-stage
RUN go test -v ./...

# Application binary into a lean image
FROM scratch AS build-release-stage

WORKDIR /app

COPY --from=build-stage /server /server

# Bind TCP port
EXPOSE 8090

# Run
ENTRYPOINT [ "/server" ]