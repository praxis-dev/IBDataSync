FROM golang:1.18.1-alpine as builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/cmd/server

RUN go build -o /out/main .

FROM golang:1.18.1-alpine

RUN apk --no-cache add ca-certificates

RUN go install github.com/pilu/fresh@latest

WORKDIR /app

COPY . .

EXPOSE 8080

CMD ["sh", "-c", "cd /app/cmd/server && fresh"]
