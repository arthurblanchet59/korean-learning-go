FROM golang:1.22-alpine AS build

WORKDIR /src

COPY go.work ./
COPY apps/backend/go.mod ./apps/backend/go.mod
COPY apps/backend/go.sum ./apps/backend/go.sum
COPY packages/core/go.mod ./packages/core/go.mod

WORKDIR /src/apps/backend
RUN go mod download

WORKDIR /src
COPY packages/core ./packages/core
COPY apps/backend ./apps/backend

RUN go build -o /out/korean-learning-backend ./apps/backend

FROM alpine:3.20

RUN adduser -D -H appuser
USER appuser

WORKDIR /app
COPY --from=build /out/korean-learning-backend ./korean-learning-backend

EXPOSE 8080

CMD ["./korean-learning-backend"]
