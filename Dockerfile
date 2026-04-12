FROM golang:1.22-alpine AS build

WORKDIR /src

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /taskflow \
    ./cmd/server

FROM alpine:3.19 AS runtime

RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /app

COPY --from=build /taskflow ./taskflow
COPY backend/config/ ./config/
COPY backend/data/seed ./data/seed/

EXPOSE 4000

CMD ["./taskflow"]
