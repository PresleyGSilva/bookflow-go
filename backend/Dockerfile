FROM golang:1.27-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bookflow ./cmd/api

FROM alpine:3.20
RUN adduser -D -H appuser && apk add --no-cache ca-certificates wget
USER appuser
WORKDIR /app
COPY --from=build /bookflow /app/bookflow
COPY migrations ./migrations
COPY openapi.yaml ./openapi.yaml
EXPOSE 8085
ENV PORT=8085
CMD ["/app/bookflow"]
