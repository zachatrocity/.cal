FROM golang:1.21-alpine

# Install git and openssh
RUN apk add --no-cache git openssh-client

# Set up SSH for git
RUN mkdir -p /root/.ssh && \
    chmod 700 /root/.ssh && \
    ssh-keyscan github.com > /root/.ssh/known_hosts

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o /usr/local/bin/dotcal ./cmd/server

# Create directory for git repository
RUN mkdir -p /app/repo

ENTRYPOINT ["dotcal"]
