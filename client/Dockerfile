# Use the official Golang image to create a build artifact.
FROM golang:1.19 as builder

# Copy local code to the container image.
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Build the command inside the container.
RUN go build -o /client ./client/main.go

# Use a minimal image to run the client binary.
FROM gcr.io/distroless/base-debian10

# Copy the client binary.
COPY --from=builder /client /client

# Run the client binary.
CMD ["/client"]
