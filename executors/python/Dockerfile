# Python Code Executor for Scribble
# Executes user-submitted Python code in an isolated, sandboxed environment

FROM python:3.11-slim

# Set working directory
WORKDIR /executor

# Install psutil for memory measurement
# Keep the image minimal for security
RUN pip install --no-cache-dir psutil==5.9.6

# Copy the executor runner script
COPY runner.py /executor/runner.py

# Create a non-root user for code execution (security best practice)
RUN useradd -m -u 1000 coderunner && \
    chown -R coderunner:coderunner /executor

# Switch to non-root user
USER coderunner

# Set resource limits via environment (enforced by Kubernetes)
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1

# Entry point: run the executor script
CMD ["python", "/executor/runner.py"]
