# Stage 1: Build Cloud (Go)
FROM docker.io/library/golang:1.25-bookworm AS builder-cloud

WORKDIR /app
COPY cloud/go.mod cloud/go.sum ./
RUN go mod download

COPY cloud/ .
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server

# Stage 2: Build Chat (Node.js)
FROM docker.io/library/node:20-bookworm AS builder-chat

WORKDIR /app
COPY chat/package.json chat/package-lock.json* ./
RUN npm install
RUN npm install lightningcss@latest @tailwindcss/oxide@latest

COPY chat/ .
ENV GITHUB_PAGES=true
ENV NEXT_TELEMETRY_DISABLED=1
ENV NEXT_DEPLOYMENT_TYPE=docker
RUN npm run build

# Stage 3: Runtime
FROM docker.io/library/node:20-bookworm AS runner

# Install Go runtime for cloud service
RUN apt-get update && apt-get install -y ca-certificates wget

WORKDIR /app

# Copy cloud binary
COPY --from=builder-cloud /app/server ./server

# Copy chat build output
COPY --from=builder-chat /app/public ./public
COPY --from=builder-chat /app/.next/standalone ./chat
COPY --from=builder-chat /app/.next/static ./chat/.next/static

# Create startup script
RUN echo '#!/bin/sh\n\
echo "Starting Cloud service..."\n\
./server &\n\
echo "Starting Chat service..."\n\
cd chat && node server.js\n' > /start.sh && chmod +x /start.sh

EXPOSE 8080 3000

CMD ["/start.sh"]
