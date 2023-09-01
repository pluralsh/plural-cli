#!/bin/sh

session="plural-workspace"

# ensure necessary env vars are populated
if [ -f /home/plural/.env ]; then
  source /home/plural/.env
fi

until abduco | grep -q "\*.*$session"; do sleep 0.01; done && printf "\e[?1049l" | abduco -a "$session" &
abduco -A $session /home/plural/boot.sh
