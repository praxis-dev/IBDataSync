FROM golang:1.18.1-alpine as builder

WORKDIR /app

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
WORKDIR /app/cmd/server
RUN go build -o /out/main .

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /out/main .

# Install Air for live reloading
RUN apk add --no-cache curl
RUN curl -fLo ./air https://github.com/cosmtrek/air/releases/download/v1.40.4/air_1.40.4_linux_amd64
RUN chmod +x ./air

EXPOSE 8080
CMD ["./air"]
