#!/bin/sh

session="workspace"

# ensure necessary env vars are populated
if [ -f /home/plural/.env ]; then
  source /home/plural/.env
fi

tmux start
tmux has-session -t $session 2>/dev/null

if [ $? != 0 ]; then
  tmux new-session -c ~/workspace -s $session /home/plural/boot.sh
fi

# Attach to created session
tmux attach-session -d -t $session