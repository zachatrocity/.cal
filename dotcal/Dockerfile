FROM golang:1.24-alpine

# Install required packages
RUN apk add --no-cache git openssh-client dcron

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

# Copy and set up entrypoint script
COPY entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/entrypoint.sh

ENTRYPOINT ["/usr/local/bin/entrypoint.sh"]
