#!/bin/bash
# this will generate a self signed certificate to be used with Shorty

YEARS=1
DAYS=$(( 365 * $YEARS ))

# gen priv key
openssl req -new -sha256 -key shorty.key -out shorty.csr
# gen cert
openssl x509 -req -sha256 -in shorty.csr -signkey shorty.key -out shorty.crt -days 3650

if [ ! -f shorty.key ] || [ ! -f shorty.crt ]; then
    echo "Error generating public cert and private key." 1>&2
    exit 1
fi

