#!/bin/sh

cd "$(dirname "$0")"
./launcher -e age2 --gameConfig resources/config.aoe2.toml
if [ $? -eq 0 ]; then
  echo "Program finished successfully, closing in 10 seconds..."
  sleep 10
else
  echo "Program finished with errors..."
  read -p "Press any key to exit..."
fi
