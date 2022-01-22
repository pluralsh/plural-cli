#!/bin/sh

session="workspace"
tmux start
tmux has-session -t $session 2>/dev/null

if [ $? != 0 ]; then
  tmux new-session -c /root/workspace -s $session zsh
fi

# Attach to created session
tmux attach-session -t $session