#!/bin/bash

started_at=$(date '+%Y%m%d %H:%M')
echo "starting task" $started_at
echo "haj!"
sleep 10;
echo "ending task", $started_at, $(date '+%Y%m%d %H:%M')
exit 1;