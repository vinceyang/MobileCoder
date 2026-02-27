#!/bin/bash
set -e

echo "=== Initializing server for Docker deployment ==="

# Check root permission
if [ "$EUID" -ne 0 ]; then
    echo "Please run as root"
    exit 1
fi

# Install Docker
echo "Installing Docker..."
curl -fsSL https://get.docker.com | sh

# Start Docker
echo "Starting Docker..."
systemctl start docker
systemctl enable docker

# Install Docker Compose
echo "Installing Docker Compose..."
curl -L "https://github.com/docker/compose/releases/download/v2.24.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Create deployment directory
echo "Creating deployment directory..."
mkdir -p /opt/agentapi

# Configure firewall (optional)
echo "Configuring firewall..."
# ufw allow 8080/tcp  # Cloud API
# ufw allow 3001/tcp   # Chat H5

echo "=== Initialization complete ==="
echo "Next steps:"
echo "1. Upload deployment files to /opt/agentapi/"
echo "2. Run: cd /opt/agentapi && docker-compose up -d"
