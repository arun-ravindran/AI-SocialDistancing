# Dockerfile for webserver
# TO DO: Reduce image size with multi stage build
FROM golang:alpine

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
	GITHUB_CLIENT_ID=a785f876ef176946b80c \
	GITHUB_CLIENT_SECRET=fac9727eaf597244e9ff5ef5bc0375a65cc37137

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container
COPY . .

# Copy html directory
ADD templates ./templates

# Build the application
RUN go build -o main .

# Export necessary port
EXPOSE 8080

# Command to run when starting the container
ENTRYPOINT ["/build/main"]

# For Debug using attach shell
#ENTRYPOINT ["tail", "-f", "/dev/null"]
