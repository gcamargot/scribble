# JavaScript executor for Scribble
# Uses Node.js 20 LTS (Alpine) for small image size
FROM node:20-alpine

# Create non-root user for security
RUN adduser -D -u 1000 coderunner

# Set working directory
WORKDIR /app

# Copy runner script
COPY runner.js /app/runner.js

# Make runner executable
RUN chmod +x /app/runner.js

# Switch to non-root user
USER coderunner

# Set Node.js options for security
ENV NODE_OPTIONS="--max-old-space-size=256"

# Entry point
CMD ["node", "/app/runner.js"]
