FROM golang:1.22-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /app/apps/subgraph-intervention
RUN go build -o main .

EXPOSE 8082

CMD ["./main"]
