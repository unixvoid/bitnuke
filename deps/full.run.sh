#!/bin/sh
VER_NO="v0.10.4 PRE_RELEASE:$DIFF"

echo "daemonize yes" > /redis.conf
echo "dbfilename dump.rdb" >> /redis.conf
echo "dir /redisbackup/" >> /redis.conf
echo "save 30 1" >> /redis.conf

redis-server /redis.conf
nginx

echo -e ""
echo -e "\e[36m                    ________     \e[39m"
echo -e "\e[36m   _______________ |        |    \e[39m"
echo -e "\e[36m  /               \ \      /     \e[39m"
echo -e "\e[36m |                 |  ----       \e[39m"
echo -e "\e[36m |                 |  ----       \e[39m"
echo -e "\e[36m  \_______________/ /      \     \e[39m"
echo -e "\e[36m                   |________|    \e[39m"
echo -e "\e[0m-------- \e[31mbitnuke\e[39m --------------\e[0m"
echo -e "\e[0m  :: \e[31m$VER_NO\e[39m ::  \e[0m"

bitnuke $@
