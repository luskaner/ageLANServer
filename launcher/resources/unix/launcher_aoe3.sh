#!/bin/sh

cd "$(dirname "$0")"
./launcher -e age3 -c resources/config.aoe3.toml
read -p "Press any key to exit..."