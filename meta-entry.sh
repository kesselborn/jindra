#!/bin/sh -x
mkdir -p /root/.ssh
cat /mnt/ssh/authorized_keys > /etc/authorized_keys/root
/entry.sh "$@" &
pid=$!

echo "waiting for stages to finish"
while test -e ${STAGES_RUNNING_SEMAPHORE}; do printf "."; sleep 1; done

kill -9 $pid
