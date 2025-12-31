# Go executor for Scribble
# Uses Go Alpine for compilation
FROM golang:1.23-alpine

# Install Python for test runner and other tools
RUN apk add --no-cache python3 coreutils

# Create non-root user for security
RUN adduser -D -u 1000 coderunner

# Set working directory
WORKDIR /app

# Copy runner script
COPY runner.sh /app/runner.sh
RUN chmod +x /app/runner.sh

# Create temp directory with proper permissions
RUN mkdir -p /tmp && chmod 1777 /tmp

# Switch to non-root user
USER coderunner

# Set Go environment
ENV GOCACHE=/tmp/go-cache
ENV GOPATH=/tmp/go

# Entry point
CMD ["/app/runner.sh"]
