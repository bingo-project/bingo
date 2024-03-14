FROM alpine:latest
LABEL maintainer="<brooksyang@outlook.com>"

WORKDIR /opt/bingo

COPY _output/cert /opt/bingo/cert
COPY _output/bin/bingo-apiserver /opt/bingo/bin/
COPY configs/*.yaml /etc/bingo/

COPY build/docker /opt/bingo/

ENTRYPOINT ["/opt/bingo/bin/bingo-apiserver"]
