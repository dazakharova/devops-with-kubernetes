FROM golang AS build
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/broadcaster ./main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=build /out/broadcaster /broadcaster
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/broadcaster"]
