FROM node:20-alpine AS ui-builder
WORKDIR /app/ui
COPY ui/package.json ui/package-lock.json ./
RUN npm ci --silent
COPY ui/ .
RUN npm run build

FROM golang:1.23-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=ui-builder /app/internal/ui/static ./internal/ui/static
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /drift ./cmd/drift

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /drift /usr/local/bin/drift
ENTRYPOINT ["drift"]
