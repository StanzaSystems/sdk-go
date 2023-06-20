#!/bin/bash

if [ -x $(which vegata) ]
then
    # TODO: make duration and rate overrideable
    duration="60s"
    rate="10/s"

    url="http://localhost:3000/account/sre-is-msg"
    cmd="vegeta attack -duration=${duration} -rate=${rate}"

    rm *.gob &>/dev/null
    echo "Running two parallel loadtests (${duration} duration ${rate} \"enterprise\", ${rate} \"free\")"
    echo "GET ${url}" | $cmd -header="X-User-Plan: free" -output=free.gob &
    echo "GET ${url}" | $cmd -header="X-User-Plan: enterprise" -output=enterprise.gob

    echo ""
    echo "ENTERPRISE:"
    vegeta report enterprise.gob | grep Success
    vegeta report enterprise.gob | grep Status
    echo ""
    echo "FREE:"
    vegeta report free.gob | grep Success
    vegeta report free.gob | grep Status
    echo ""
else
    echo "Please install vegata: https://github.com/tsenart/vegeta"
    echo "Hint: \"go install github.com/tsenart/vegeta@latest\""
fi