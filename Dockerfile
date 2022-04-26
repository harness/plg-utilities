##
## Build
##
FROM golang:1.17 AS build
WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /plg-utilities-job ./cmd/main.go

##
## Deploy
##
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/config.yaml /config.yaml
COPY --from=build /plg-utilities-job /plg-utilities-job
ENTRYPOINT ["/plg-utilities-job"]