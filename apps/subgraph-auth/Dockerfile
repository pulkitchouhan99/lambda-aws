FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/apps/subgraph-auth
RUN go build -o main .

EXPOSE 8081

CMD ["./main"]
