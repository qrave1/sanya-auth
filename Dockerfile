FROM golang:1.22-alpine AS builder
LABEL authors="kosma4"

ENV CGO_ENABLED=0 GOOS=linux
WORKDIR /auth

COPY . .
RUN go mod download

RUN go build -o app cmd/main.go

FROM alpine as runtime

COPY --from=builder /auth/app /app
CMD ["/app"]
