#!/bin/sh

if ./battle-server-manager start -e "${GAME_ID}" --logRoot="/app/logs/${GAME_ID}/${LOG_SUBFOLDER}" ${BS_MANAGER_ARGS}
then
    echo "Monitoring BattleServer.exe..."
    while pgrep -f BattleServer.exe > /dev/null; do
        sleep 10
    done
    echo "BattleServer.exe stopped."
else
    exit 1
fi