FROM golang:1.23-alpine AS builder

RUN apk add --no-cache gcc musl-dev git

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o ghloc .

FROM alpine:3.20

RUN apk add --no-cache git ca-certificates

COPY --from=builder /app/ghloc /usr/local/bin/ghloc

EXPOSE 8080

ENTRYPOINT ["ghloc"]
