FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/backend .

FROM alpine:3.22

RUN apk add --no-cache ca-certificates

WORKDIR /pb

COPY --from=build /out/backend ./backend
COPY pb_migrations ./pb_migrations

EXPOSE 8090

CMD ["./backend", "serve", "--http=0.0.0.0:8090"]
