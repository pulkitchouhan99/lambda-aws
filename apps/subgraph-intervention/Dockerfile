FROM golang:1.24-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/apps/subgraph-intervention
RUN go build -o main .

EXPOSE 8082

CMD ["./main"]
