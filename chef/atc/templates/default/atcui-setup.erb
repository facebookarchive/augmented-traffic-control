#!/bin/bash

RED="\e[0;31m"
GREEN="\e[0;32m"
NC="\e[00m"

USER="<%= node['atc']['atcui']['user'] %>"

# Allow either $USER or root to run this script.
if [ "$(whoami)" == "$USER" ] ; then
    USER=""
elif [ "$EUID" != "0" ] ; then
   echo -e "${RED}$0 must be run as root or $USER.${NC}"
   exit 1
fi

VENV="<%= File.join(node['atc']['venv']['path'], 'bin', 'activate') %>"

function p() {
    echo -e "${GREEN}${1}${NC}"
}

function managepy() {
    C="cd /var/django && . \"$VENV\" && python manage.py $@"
    if [ -z "$USER" ] ; then
        /bin/bash -c "$C"
    else
        sudo -u "$USER" /bin/bash -c "$C"
    fi
}

p "Migrating DB"
managepy migrate

p "Loading Samples"
managepy loaddata sample

p "Generating static files"
managepy collectstatic --noinput
