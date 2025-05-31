FROM golang:1.22-alpine AS builder

WORKDIR /app
# Download dependencies
COPY go.mod go.sum /
RUN go mod download

# Add source code
COPY . .
RUN go build -o ashno-onepay ./cmd/main.go

# Multi-Stage production build
FROM alpine AS production
RUN apk add --no-cache tzdata

WORKDIR /app
# Retrieve the binary from the previous stage
COPY --from=builder /app/ashno-onepay /app/ashno-onepay
# Expose port
EXPOSE 8000
# Set the binary as the entrypoint of the container
ENTRYPOINT ["./ashno-onepay"]