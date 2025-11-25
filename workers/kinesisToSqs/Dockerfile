FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/workers/kinesisToSqs
RUN go build -o main .

CMD ["./main"]