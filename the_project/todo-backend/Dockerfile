# Build stage
FROM golang:alpine AS builder

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -ldflags='-s' -o todo-backend ./cmd/todo-backend

# Runtime stage
FROM scratch AS runtime

WORKDIR /app

COPY --from=builder /app/todo-backend .

ENV PORT=4001

ENTRYPOINT ["./todo-backend"]
