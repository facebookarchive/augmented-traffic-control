#!/bin/bash

PLATFORM=$(uname -s)

if [[ "${PLATFORM}" == 'Darwin' ]]; then
    VBOXMANAGE='/Applications/VirtualBox.app/Contents/MacOS/VBoxManage'
else
    VBOXMANAGE='/usr/lib/virtualbox/VBoxManage'
fi

function get_genymotion_uids () {
    ${VBOXMANAGE} list vms | grep 'API' | awk -F'[{}]' '{print $2}'
}

function get_atc_uid () {
    ${VBOXMANAGE} list vms | egrep 'atc(centos|ubuntu)' | awk -F'[{}]' '{print $2}'
}

get_genymotion_uids
get_atc_uid
