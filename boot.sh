#!/bin/bash

if [ -f /home/plural/.env ]; then
  source /home/plural/.env
fi

cd ~/workspace || echo "could not check out workspace repo, ensure it exists and git permissions are correct"
zsh