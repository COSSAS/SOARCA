FROM scratch
LABEL MAINTAINER Author maarten de kruijf, jan-paul konijn

ARG BINARY_NAME=soarca
ARG VERSION

COPY bin/${BINARY_NAME}-${VERSION}-linux-amd64 /bin/soarca

WORKDIR /bin

EXPOSE 8080

CMD ["./soarca"]