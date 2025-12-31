# Production Dockerfile for Frontend (React + Vite)
# Multi-stage build: Build with Node, serve with nginx

# Stage 1: Build the React application
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package files
COPY package.json package-lock.json* ./

# Install dependencies
RUN npm ci --prefer-offline || npm install

# Copy source code
COPY . .

# Build for production with Vite
# This creates optimized static files in dist/
RUN npm run build

# Stage 2: Serve with nginx
FROM nginx:1.25-alpine

# Copy built static files from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy custom nginx configuration for SPA routing
# This ensures that all routes fall back to index.html for client-side routing
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80
EXPOSE 80

# Health check to ensure nginx is serving content
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost/ || exit 1

# Start nginx in foreground (daemon off is required for Docker)
CMD ["nginx", "-g", "daemon off;"]
