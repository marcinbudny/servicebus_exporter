FROM golang:1.12-stretch as build

WORKDIR /build
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-extldflags "-static"' -o app

FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
COPY --from=build /build /
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
EXPOSE 9580
ENTRYPOINT [ "/app" ]