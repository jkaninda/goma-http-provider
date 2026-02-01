########################
# Builder Stage
########################
FROM golang:1.25.5 AS build

WORKDIR /app

# Copy source code
COPY . .

# Download Go dependencies
RUN go mod download

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/goma cmd/main.go

########################
# Final Stage
########################
FROM alpine:3.22.0

ENV TZ=UTC

# Install runtime dependencies and set up directories
RUN apk --update --no-cache add tzdata ca-certificates curl

# Copy built binary
COPY --from=build /app/goma /usr/local/bin/goma
RUN chmod a+x /usr/local/bin/goma && ln -s /usr/local/bin/goma /goma

# Set working directory
WORKDIR /app
# Expose HTTP Port
EXPOSE 8080

ENTRYPOINT ["/goma"]