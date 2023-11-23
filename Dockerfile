# Use an official Go runtime as a parent image
FROM golang:1.18

# Set the working directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files to download the dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go app
# Note the ./cmd path to build the main.go file within the cmd directory
RUN go build -o redislock ./cmd

# Run the binary program produced by `go build`
CMD ["./redislock"]
