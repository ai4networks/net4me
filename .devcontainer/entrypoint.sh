#!/usr/bin/env bash
set -o errexit
set -o nounset
set -o pipefail
if [[ "${TRACE-0}" == "1" ]]; then
    set -o xtrace
fi

function start_openvswitch_service {
    echo "*** Starting OpenvSwitch service..."
    service openvswitch-switch start > /dev/null 2>&1
    ovs-vswitchd --pidfile --detach > /dev/null 2>&1
    ovs-vsctl set-manager ptcp:6640 > /dev/null 2>&1
    echo "*** OpenvSwitch service started ✅"
}

start_openvswitch_service
