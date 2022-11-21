#!/bin/sh

session="plural-workspace"

# ensure necessary env vars are populated
if [ -f /home/plural/.env ]; then
  source /home/plural/.env
fi

screen -xR -S $session /home/plural/boot.sh
