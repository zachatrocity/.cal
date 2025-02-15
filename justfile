# List available recipes
default:
    @just --list

# Run Go tests with coverage
test:
    cd dotcal && go test ./... -coverprofile=coverage.out -v
    cd dotcal && go tool cover -func=coverage.out

# Build Docker container
build:
    docker-compose build

# Start the container
start:
    docker-compose up -d

# Stop the container
stop:
    docker-compose down

# View container logs
logs:
    docker-compose logs -f

# Connect to container shell
shell:
    docker-compose exec dotcal sh

# Clean up Docker resources
clean:
    docker-compose down --rmi all --volumes

# Build and start the container
up: build start

# Stop and remove the container, then rebuild and start
rebuild: stop build start

# Run tests and show coverage in browser
coverage:
    cd dotcal && go test ./... -coverprofile=coverage.out
    cd dotcal && go tool cover -html=coverage.out

# Create and push a new release tag (usage: just release 0.0.2)
release version:
    git tag v{{version}}
    git push origin v{{version}}
