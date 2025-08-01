FROM golang:1.22 AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o auto-namespace-intrumentation-injector main.go

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /app/auto-namespace-intrumentation-injector .
USER nonroot:nonroot
ENTRYPOINT ["/auto-namespace-intrumentation-injector"]
