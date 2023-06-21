#!/bin/bash

function spinner() {
    local pid=$! # Process Id of the previous running command
    local spin='-\|/'
    local charwidth=1
    local i=0
    tput civis # cursor invisible
    while kill -0 $pid 2>/dev/null; do
        local i=$(((i + $charwidth) % ${#spin}))
        printf "%s" "${spin:$i:$charwidth}"
        echo -en "\033[1D" # cursor back one
        sleep .1
    done
    tput cnorm
    wait $pid # capture exit code
}

if [ -x $(which vegata) ]
then
    duration="30s"
    rate="10/s"

    # https://ca.slack-edge.com/T029RQSE6-U014D01R5U6-1efa53dd27be-72
    url="http://localhost:3000/test/aHR0cHM6Ly9jYS5zbGFjay1lZGdlLmNvbS9UMDI5UlFTRTYtVTAxNEQwMVI1VTYtMWVmYTUzZGQyN2JlLTcy"
    cmd="vegeta attack -duration=${duration} -rate=${rate}"

    rm *.gob &>/dev/null
    echo -e "Running two parallel loadtests (\033[38;5;11m${duration} duration\e[0m, \e[32m${rate} \"enterprise\"\e[0m and \e[31m${rate} \"free\"\e[0m)"
    echo "GET ${url}" | $cmd -header="X-User-Plan: free" -output=free.gob &
    echo "GET ${url}" | $cmd -header="X-User-Plan: enterprise" -output=enterprise.gob &
    spinner

    echo ""
    echo -e "\e[32mENTERPRISE\e[0m"
    vegeta report enterprise.gob | grep Success
    vegeta report enterprise.gob | grep Status
    echo ""
    echo -e "\e[31mFREE\e[0m"
    vegeta report free.gob | grep Success
    vegeta report free.gob | grep Status
    echo ""
    tput cnorm # reset cursor
else
    echo "Please install vegata: https://github.com/tsenart/vegeta"
    echo "Hint: \"go install github.com/tsenart/vegeta@latest\""
fi