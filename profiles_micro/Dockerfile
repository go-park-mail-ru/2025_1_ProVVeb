FROM golang:1.23-alpine

WORKDIR /backend
COPY go.mod go.sum ./
RUN go mod download

COPY . .

WORKDIR /backend/profiles_micro

CMD ["go", "run", "main.go"]
