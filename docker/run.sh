#!/bin/sh

echo "daemonize yes" > /redis.conf
redis-server /redis.conf
/bitnuke
