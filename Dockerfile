FROM alpine

RUN apk update && apk add redis

COPY bitnuke /bitnuke
COPY upload /upload
COPY run.sh /run.sh

RUN echo "daemonize yes" > /redis.conf

CMD ["/run.sh"]
