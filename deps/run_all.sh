#!/bin/ash

# start redis server
redis-server /redis.conf | sed "s/^/[redis] /" &

# start bitnuke
bitnuke | sed "s/^/[bitnuke] /" &

# start nginx
nginx | sed "s/^/[nginx] /"
