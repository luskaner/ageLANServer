#!/bin/sh

cd "$(dirname "$0")"
./launcher -e age1 --gameConfig resources/config.aoe1.toml
if [ $? -eq 0 ]; then
  echo "Program finished successfully, closing in 10 seconds..."
  sleep 10
else
  echo "Program finished with errors..."
  echo "Press any key to exit..."
  read dummy
fi
