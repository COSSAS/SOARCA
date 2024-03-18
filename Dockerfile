FROM golang:alpine as builder
RUN apk update && apk upgrade && apk add --no-cache ca-certificates
RUN update-ca-certificates

FROM scratch
LABEL MAINTAINER Author maarten de kruijf, jan-paul konijn

ARG BINARY_NAME=soarca
ARG VERSION

COPY bin/${BINARY_NAME}-${VERSION}-linux-amd64 /bin/soarca
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

WORKDIR /bin

EXPOSE 8080

CMD ["./soarca"]