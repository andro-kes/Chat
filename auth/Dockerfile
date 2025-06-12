FROM golang:1.23-alpine
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . ./auth
RUN go build -o auth ./auth/*.go
ENTRYPOINT [ "/app/auth" ]