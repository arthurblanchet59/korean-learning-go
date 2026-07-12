FROM golang:1.22-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
COPY packages/core/go.mod ./packages/core/go.mod
RUN go mod download

COPY . .
RUN go build -o /out/korean-learning-api .

FROM alpine:3.20

WORKDIR /app
COPY --from=build /out/korean-learning-api ./korean-learning-api
COPY docs ./docs
RUN adduser -D -H appuser \
    && mkdir -p /app/data /app/logs \
    && chown -R appuser:appuser /app
USER appuser

EXPOSE 8080

CMD ["./korean-learning-api"]
