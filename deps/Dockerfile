FROM alpine

ENV DIFF=<DIFF>
RUN apk update && apk add redis
RUN rm -rf /var/cache/apk/*

COPY bitnuke /bitnuke
COPY run.sh /run.sh

CMD ["/run.sh"]
