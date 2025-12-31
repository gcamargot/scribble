# C++ executor for Scribble
# Uses Alpine with GCC for compilation
FROM alpine:3.19

# Install build tools and Python for test runner
RUN apk add --no-cache g++ python3 coreutils

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
