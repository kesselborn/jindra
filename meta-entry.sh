#!/bin/sh -x
mkdir -p /root/.ssh
cat /mnt/ssh/authorized_keys > /etc/authorized_keys/root
/entry.sh "$@" &
pid=$!

while test -e /jindra/semaphores/stages-running; do sleep 1; done

kill -9 $pid
