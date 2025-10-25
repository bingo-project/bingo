FROM alpine:3.22
LABEL maintainer="<brooksyang@outlook.com>"

WORKDIR /opt/bingo

# Tools
RUN apk add curl

# Timezone
# RUN apk --no-cache add tzdata && ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
#      echo "Asia/Shanghai" > /etc/timezone \

COPY _output/platforms/linux/amd64/bingo-bot bin/

EXPOSE 8080

ENTRYPOINT ["/opt/bingo/bin/bingo-bot"]
