# Use the official Golang image as a parent image
FROM golang:1.20

# Set the working directory inside the container
WORKDIR /app

# Copy the Go application source code into the container
COPY . .

# Download go dependencies
RUN go get -d -v ./...

# Build the Go application
RUN go build -o wildlife-preserve-app main.go

# Specify the entry point for the container
CMD ["./wildlife-preserve-app"]

# Create a directory for the mounted volume
RUN mkdir -p /animals-storage

# Expose the port your application listens on (if applicable)
# EXPOSE 8080

# Add a health check or other instructions as needed

# Add the volume mount point to the CMD so it's accessible
CMD ["./wildlife-preserve-app", "--animals-storage=/animals-storage"]