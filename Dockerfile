FROM golang:1.27-bookworm
ENV CGO_ENABLED=0 GOFLAGS=-mod=mod
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
CMD ["go", "test", "./..."]
