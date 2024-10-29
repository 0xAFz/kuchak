# Kuchak

A high-performance URL shortening service built with Go, designed to handle high-throughput URL shortening and redirection with robust caching and rate limiting.

## Overview

This URL shortener service provides a reliable and scalable solution for creating shortened URLs. Built with Go, it leverages Redis for caching and rate limiting, and PostgreSQL for persistent storage, ensuring both high performance and data durability.

## Prerequisites

### Development Environment
- Go version 1.22.5
- Docker Engine
- Docker Compose Plugin
- Make (for development convenience)

### Production Environment
- Docker Engine
- Docker Compose Plugin

## Local Development Setup

### 1. Clone the Repository
```bash
git clone https://github.com/0xAFz/kuchak.git
cd kuchak/
```

### 2. Environment Configuration
```bash
# Copy the example environment file
cp .env.example .env

# Edit the environment variables according to your setup
vim .env
```

### 3. Start Required Services
The project requires PostgreSQL and Redis, which are configured in the Docker Compose file.

```bash
# Start PostgreSQL and Redis containers
docker compose up -d

# Verify containers are running
docker compose ps
```

### 4. Run the Application
There are two ways to run the application locally:

```bash
# Option 1: Using Make
make run

# Option 2: Direct Go command
go run main.go
```

### 5. Verify Installation
The application should now be running on `http://localhost:1323` (or your configured port)

## Production Deployment

### 1. Install Docker
```bash
# Install Docker using the official installation script
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Start and enable Docker service
sudo systemctl start docker
sudo systemctl enable docker
```

### 2. Environment Configuration
```bash
# Copy the production environment file
cp .env.prod .env

# Configure production environment variables
vim .env
```

### 3. Deploy with Docker Compose
```bash
# Build and start the application stack
docker compose -f prod.compose.yml up --build -d

# Verify all services are running
docker compose -f prod.compose.yml ps
```

### 4. Verify Deployment
```bash
# Check application logs
docker compose -f prod.compose.yml logs -f app

# Check application health
curl http://sub.domain.tld/healthz
```

## Common Operations

### View Logs
```bash
# Development
docker compose logs -f

# Production
docker compose -f prod.compose.yml logs -f
```

### Stop Services
```bash
# Development
docker compose down

# Production
docker compose -f prod.compose.yml down
```

### Database Operations
```bash
# Connect to PostgreSQL
docker compose exec postgres psql -U <username> -d <database>

# Backup Database
docker compose exec postgres pg_dump -U <username> <database> > backup.sql
```

### Redis Operations
```bash
# Connect to Redis CLI
docker compose exec redis redis-cli
```

## Troubleshooting

### Common Issues

1. **Port Conflicts**
   - Error: `bind: address already in use`
   - Solution: Check for processes using required ports
     ```bash
     sudo lsof -i :<port>
     ```

2. **Database Connection Issues**
   - Verify PostgreSQL container is running
   - Check environment variables
   - Ensure network connectivity between services

3. **Redis Connection Issues**
   - Verify Redis container is running
   - Check Redis password in environment variables
   - Verify network connectivity

### Getting Help
For additional help:
- Check the project's GitHub issues
- Review the error logs using `docker compose logs`
- Contact the development team

## Security Considerations

1. **Environment Variables**
   - Never commit `.env` files to version control
   - Use strong passwords in production
   - Regularly rotate credentials

2. **Network Security**
   - Use internal Docker networks when possible
   - Limit exposed ports
   - Configure firewalls appropriately

3. **Updates**
   - Regularly update Docker images
   - Keep Go dependencies updated
   - Monitor security advisories
