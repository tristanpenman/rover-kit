FROM golang:1.25.8

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

ENV PATH="/app/bin:${PATH}"
