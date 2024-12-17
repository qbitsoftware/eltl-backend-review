FROM golang:1.22.5-alpine AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o /app/app ./cmd/api/main.go

FROM alpine:3.20 AS run

COPY --from=build /app /app

WORKDIR /app

EXPOSE 8080

CMD ["/app/app"]
