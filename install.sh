#!/bin/bash

if [ $UID -ne "0" ]; then
    echo "error: must be root" 1>&2
    exit 1
fi

./build.sh --target-current

mkdir -p /var/lib/mirror-sync/
mkdir -p /opt/mirror-sync

id mirror-sync 1>/dev/null 2>&1
if [ $? -ne 0 ]; then
    useradd -r -M -d /var/lib/mirror-sync/ mirror-sync
fi

cp ./build/mirrorsyncd /opt/mirror-sync/mirrorsyncd
cp ./build/mirrorsync /usr/local/bin/mirrorsync

chmod ugo+x /usr/local/bin/mirrorsync
chmod 740 /opt/mirror-sync/mirrorsyncd

chown -R mirror-sync:mirror-sync /opt/mirror-sync
chown -R mirror-sync:mirror-sync /var/lib/mirror-sync

cat <<EOF > /etc/systemd/system/mirrorsync.service
[Unit]
Description=mirrorsync daemon
After=network.target

[Service]
WorkingDirectory=/var/lib/mirror-sync
ExecStart=/opt/mirror-sync/mirrorsyncd
User=mirror-sync
Group=mirror-sync

[Install]
WantedBy=multi-user.target
EOF

systemctl daemon-reload

echo "done installing!"