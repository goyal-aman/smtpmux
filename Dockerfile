# Build Stage
FROM golang:1.24.6-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy


COPY . .
RUN go build -o /app/plugins/round_robin/round-robin-plugin /app/plugins/round_robin/main.go
RUN go build -o smtpmux .

# Run Stage
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/smtpmux .
COPY --from=builder /app/config.yaml .
COPY --from=builder /app/plugins/ /app/plugins/
# Copy the Starlark script if it's not embedded or if you want it to be editable

EXPOSE 1020

CMD ["./smtpmux"]
