FROM golang:1.18.1-alpine as builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/cmd/server
RUN go build -o /out/main .

FROM golang:1.18.1-alpine

RUN apk --no-cache add ca-certificates curl

WORKDIR /root/

COPY --from=builder /app .

RUN curl -fLo ./air https://github.com/cosmtrek/air/releases/download/v1.40.4/air_1.40.4_linux_amd64 && chmod +x ./air

COPY .air.toml .

EXPOSE 8080

CMD ["./air"]
