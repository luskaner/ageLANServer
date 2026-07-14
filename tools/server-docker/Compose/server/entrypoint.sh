#!/bin/sh

LOG_SUBFOLDER=$(date +"%Y-%m-%dT%H-%M-%S")
exec ./server -e $GAME_ID --log --flatLog --logRoot=/app/logs/server/$GAME_ID/$LOG_SUBFOLDER $SERVER_ARGS