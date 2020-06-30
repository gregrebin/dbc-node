#!/bin/bash

PRIVKEY="$1"
SIGNATURE="$2"
FILE="$3"

openssl dgst -sha256 -sign "$PRIVKEY" -out "$SIGNATURE" "$FILE"
