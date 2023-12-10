FROM golang:alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN go build ./cmd/ha-proxy-go

FROM alpine:3 as runtime

RUN adduser --disabled-password --no-create-home app

USER app
WORKDIR /app

COPY --from=build /app/ha-proxy-go .

EXPOSE 3000
ENTRYPOINT ["/app/ha-proxy-go"]
