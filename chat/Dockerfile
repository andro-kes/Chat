FROM golang:1.23-alpine
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . .
COPY web /app/web
RUN go build -o chat ./cmd/main.go
ENTRYPOINT [ "/app/chat" ]