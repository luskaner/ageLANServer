#!/bin/sh

cd "$(dirname "$0")"
./genCert
echo "Press any key to exit..."
read dummy