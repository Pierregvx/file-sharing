# Use the official Golang image to create a build artifact.
FROM golang:1.19 as builder

# Copy local code to the container image.
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . ./

# Build the command inside the container.
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /client ./client/main.go

# Use a minimal alpine image
FROM alpine:latest

# Copy the binary
COPY --from=builder /client /client

# Make it executable
RUN chmod +x /client

# Use environment variables to specify default behavior
ENV OPERATION=upload
ENV FILEPATHS="demo/file1.txt,demo/file2.txt"

# Run the binary.
CMD ["/client", "-operation=$OPERATION", "-filepaths=/app/demo/file1.txt"]
