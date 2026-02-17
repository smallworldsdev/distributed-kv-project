# Official GO image
FROM golang:1.26-alpine

# Sets current directory inside container
WORKDIR /app

# Download dependencies first so if source code changes but dependencies don't,
# dependencies won't be reinstalled
COPY go.mod go.sum* ./
RUN go mod download

# Copy source files
COPY . .

RUN go build -o node ./cmd/node

EXPOSE 8000

CMD ["./node"]
