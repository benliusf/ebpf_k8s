#!/bin/bash -e

while ! pid=$(pgrep -x "pong_server") > /dev/null; do
  echo "waiting for pong_server"
  sleep 1
done

exec ./pong_ebpf -server_pid=$pid
