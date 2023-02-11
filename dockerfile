FROM golang:1.20-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./
COPY ha ./ha

RUN go build -o /ha-proxy-go

FROM alpine:3.17 as runtime

RUN adduser --disabled-password --no-create-home app

USER app
WORKDIR /app

COPY --from=build /ha-proxy-go .
COPY templates ./templates
COPY wwwroot ./wwwroot

EXPOSE 3000
ENTRYPOINT ["/app/ha-proxy-go"]
