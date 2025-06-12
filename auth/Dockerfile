FROM golang:1.23-alpine
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . .
RUN go build -o auth ./cmd/main.go
ENTRYPOINT [ "/app/auth" ]