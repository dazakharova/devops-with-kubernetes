#Build stage
FROM golang:alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY go.mod ./

RUN go mod download

COPY . .

RUN go build -ldflags='-s' -o todo ./cmd/todo

#Runtime stage
FROM scratch AS runtime

ENV PORT=8080

WORKDIR /app

COPY --from=builder /app/todo .
COPY --from=builder /app/public ./public

ENTRYPOINT ["./todo"]


