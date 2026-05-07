FROM golang:1.26

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN go build -o main ./cmd/sentinel

EXPOSE 8081

CMD ["./main"]