FROM golang:1.23-alpine AS builder

#Installing ca-certificate package
RUN apk update && apk upgrade && apk add --no-cache ca-certificates
RUN update-ca-certificates

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -o rpc-balancer .

# Adding nobody user
RUN echo "nobody:x:999:999:Nobody:/:" > /etc_passwd

# Build a small image
FROM scratch
WORKDIR /app
COPY --from=builder /build/rpc-balancer /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY config.yaml /app/config.yaml

USER nobody
# Command to run
ENTRYPOINT ["/app/rpc-balancer"]