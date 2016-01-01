FROM alpine

RUN apk update && apk add redis
RUN rm -rf /var/cache/apk/*

COPY bitnuke /bitnuke
COPY static /static
COPY run.sh /run.sh

CMD ["/run.sh"]
