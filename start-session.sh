#!/bin/sh

session="/tmp/plural-workspace"

# ensure necessary env vars are populated
if [ -f /home/plural/.env ]; then
  source /home/plural/.env
fi

dtach -A $session /home/plural/boot.sh
