FROM golang:1.23-alpine
WORKDIR /backend
COPY go.mod go.sum ./
RUN go mod download

RUN mkdir -p /backend/logs
COPY . .

WORKDIR /backend/cmd

CMD ["go", "run", "main.go"]