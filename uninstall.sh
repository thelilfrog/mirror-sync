#!/bin/bash

PURGE=false

usage() {
 echo "Usage: $0 [OPTIONS]"
 echo "Options:"
 echo " --purge    Remove also data"
}

# Function to handle options and arguments
handle_options() {
  while [ $# -gt 0 ]; do
    case $1 in
      --purge)
        PURGE=true
        ;;
      *)
        echo "Invalid option: $1" >&2
        usage
        exit 1
        ;;
    esac
    shift
  done
}

# Main script execution
handle_options "$@"

if [ $UID -ne "0" ]; then
    echo "error: must be root" 1>&2
    exit 1
fi

if [ "$PURGE" == "true" ]; then
    rm -r /var/lib/mirror-sync
    if [ $? -ne 0 ]; then
        echo "failed to remove data" 1>&2
    fi
fi

rm -r /opt/mirror-sync
if [ $? -ne 0 ]; then
    echo "failed to uninstall" 1>&2
    exit 1
fi

rm -r /usr/local/bin/mirrorsync
if [ $? -ne 0 ]; then
    echo "failed to uninstall" 1>&2
    exit 1
fi

userdel mirror-sync
if [ $? -ne 0 ]; then
    echo "failed to uninstall" 1>&2
    exit 1
fi

rm -r /etc/systemd/system/mirrorsync.service
if [ $? -ne 0 ]; then
    echo "failed to uninstall" 1>&2
    exit 1
fi

systemctl disable mirrorsync.service 1>/dev/null 2>&1
systemctl daemon-reload

echo "done uninstalling!"