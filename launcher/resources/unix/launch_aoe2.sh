#!/bin/sh

cd "$(dirname "$0")"
./launcher -e age2 -c resources/config.aoe2.toml
read -p "Press any key to exit..."