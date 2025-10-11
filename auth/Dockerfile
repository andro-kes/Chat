FROM golang:1.24-alpine
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . .
COPY web /app/web
RUN go build -o auth ./cmd/main.go
ENTRYPOINT [ "/app/auth" ]