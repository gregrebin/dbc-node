#!/bin/bash

ECPARAM="$1"
PRIVKEY="$2"
PUBKEY="$3"

openssl ecparam -name secp256k1 -out "$ECPARAM"
openssl ecparam -in "$ECPARAM" -genkey -noout -out "$PRIVKEY"
openssl ec -in "$PRIVKEY" -pubout -out "$PUBKEY"