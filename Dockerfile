FROM golang:1.19-alpine3.16 AS build
RUN apk add --no-cache musl-dev gcc git make
WORKDIR /root/build
COPY *.go go.mod go.sum Makefile /root/build/
RUN make

FROM alpine:3.16
RUN apk add --no-cache iptables socat bash sudo coreutils procps
WORKDIR /app
COPY --from=build /root/build/tcp-hole-puncher /app/
COPY run-hole-maker.sh handle-address.sh /app/
RUN mkdir -p /app/handler/ && ln -sf /app/handler/address.txt /app/address.txt

CMD ["/app/run-hole-maker.sh", "--trace"]
