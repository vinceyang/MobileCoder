#!/bin/bash
set -e

SERVER=${1:-}
if [ -z "$SERVER" ]; then
    echo "Usage: ./deploy.sh <user@server>"
    echo "Example: ./deploy.sh root@123.45.67.89"
    exit 1
fi

echo "=== Deploying to $SERVER ==="

# Export images
echo "Exporting images..."
docker save -o /tmp/agentapi-cloud.tar agentapi/cloud:latest
docker save -o /tmp/agentapi-chat.tar agentapi/chat:latest

# Upload to server
echo "Uploading to server..."
scp /tmp/agentapi-cloud.tar /tmp/agentapi-chat.tar ${SERVER}:/tmp/

# Upload config files
echo "Uploading config files..."
scp deploy/docker-compose.yml ${SERVER}:/opt/agentapi/
scp deploy/nginx.conf ${SERVER}:/opt/agentapi/

# Load images and start services on server
echo "Loading images and starting services..."
ssh $SERVER << 'EOF'
    cd /opt/agentapi

    # Load images
    docker load -i /tmp/agentapi-cloud.tar
    docker load -i /tmp/agentapi-chat.tar

    # Start services
    docker-compose up -d

    # Clean up temp files
    rm -f /tmp/agentapi-cloud.tar /tmp/agentapi-chat.tar

    echo "=== Deployment complete ==="
    docker-compose ps
EOF

echo "=== Done ==="
