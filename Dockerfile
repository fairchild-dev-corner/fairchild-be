# Use the official Go image with Alpine for a small footprint
FROM golang:1.26-alpine

# Install make
RUN apk add --no-cache make

# Set working directory
WORKDIR /app/

# Copy source files and Makefile into container
COPY . .

# Manually add log 
RUN mkdir -p /fairchild_be/log && touch /fairchild_be/log/general-loggers.log

# Build app and specify ENV
RUN make build-stage

# Default command
CMD ["make", "run-staging"]