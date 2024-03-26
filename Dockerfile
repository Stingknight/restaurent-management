# syntax=docker/dockerfile:1

FROM golang:1.19-alpine

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download


# Copy the source code including folders
COPY . .

# Build
RUN go build -o main .

# Expose port 9000
EXPOSE 9000

# Run the application
CMD ["./main", "-addr=:9000"]
