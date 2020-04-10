#!/bin/bash
#curl 127.0.0.1:9090/status
#
#curl -v 127.0.0.1:9090/cache/testkey -XPUT -dtestvalue
#
#curl 127.0.0.1:9090/cache/testkey
#
#curl 127.0.0.1:9090/status
#
#curl 127.0.0.1:9090/cache/testkey -XDELETE
#
#curl 127.0.0.1:9090/status
sleep 1
#tcp test#
if [ "$1" == 'tcp' ]; then
  ./cache-benchmark.bak -type tcp -n 100000 -r 100000 -t set -c 100
  ./cache-benchmark.bak -type tcp -n 100000 -r 100000 -t get -c 100
elif [ "$1" == 'http' ]; then
#http test#
  ./cache-benchmark.bak -type http -p 9090 -n 100000 -r 100000 -t set
  ./cache-benchmark.bak -type http -p 9090 -n 100000 -r 100000 -t get
fi
  ./cache-benchmark.bak -type redis -n 100000 -r 100000 -t set -c 100
  ./cache-benchmark.bak -type redis -n 100000 -r 100000 -t get -c 100

#redis-benchmark -c 100 -n 100000 -d 1000 -t set,get -r 100000

#pid=$(netstat -anp|grep 2333 |grep main|awk '{printf $7}'|cut -d/ -f1)
#kill -9 $pid
