# Use the official Golang image to create a build artifact.
FROM golang:1.19 as builder

# Copy local code to the container image.
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Install the PostgreSQL driver for Go
RUN go get -u github.com/lib/pq

# Build the command inside the container.
RUN go build -o /server ./server/main.go

# Use a minimal image to run the server binary.
FROM gcr.io/distroless/base-debian10

# Copy the server binary.
COPY --from=builder /server /server

# Run the server binary.
CMD ["/server"]
