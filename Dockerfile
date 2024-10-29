FROM golang:1.22.5-alpine3.20 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . . 

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o /app

FROM alpine:3.20 AS prod
WORKDIR /app
COPY --from=builder /app .
EXPOSE 1323
CMD [ "./app" ]