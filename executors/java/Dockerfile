# Java executor for Scribble
# Uses Eclipse Temurin JDK 21 (Alpine) for compilation and execution
FROM eclipse-temurin:21-jdk-alpine

# Create non-root user for security
RUN adduser -D -u 1000 coderunner

# Set working directory
WORKDIR /app

# Copy and compile runner
COPY Runner.java /app/Runner.java
RUN javac /app/Runner.java

# Create temp directory for user code compilation
RUN mkdir -p /tmp/java_compile && chown coderunner:coderunner /tmp/java_compile

# Switch to non-root user
USER coderunner

# Set Java options for security and memory limits
ENV JAVA_OPTS="-Xmx256m -XX:+UseSerialGC"

# Entry point
CMD ["java", "-Xmx256m", "-XX:+UseSerialGC", "-cp", "/app", "Runner"]
