FROM alpine

RUN apk update && apk add redis
RUN rm -rf /var/cache/apk/*

COPY bitnuke /bitnuke
COPY upload /upload
COPY run.sh /run.sh

CMD ["/run.sh"]
