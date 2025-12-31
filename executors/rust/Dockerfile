# Rust executor for Scribble
# Uses Alpine with Rust for compilation
FROM rust:1.75-alpine

# Install Python for test runner and other tools
RUN apk add --no-cache python3 coreutils musl-dev

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

# Entry point
CMD ["/app/runner.sh"]
