# Use the offical Golang image to create a build artifact.
# This is based on Debian and sets the GOPATH to /go.
# https://hub.docker.com/_/golang
FROM golang:1.15-alpine as builder

RUN apk update

WORKDIR /notification-manager
COPY ../notification-manager .

# Copy local code to the container image.
WORKDIR /app
COPY . .

# Fetch dependencies first; they are less susceptible to change on every build
# and will therefore be cached for speeding up the next build.
RUN go mod download 2> /dev/null

# CGO_ENABLED=0 == Don't depend on libc (bigger but more independent binary)
# installsuffix == Cache dir for non cgo build files
RUN env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -installsuffix 'static' -o main

FROM scratch
WORKDIR /app

# Import the Certificate-Authority certificates for enabling HTTPS.

COPY --from=builder /app/main .
COPY --from=builder /app/.env .

CMD ["./main"]
