#!/usr/bin/env bash
go build -o /usr/local/bin/ecobeehvacmode
if [ -f ecobeehvacmode ];then
  cp ecobeehvacmode /usr/local/bin/ecobeehvacmode
else
  echo "Missing ecobeehvacmode binary"
  exit 1
fi
if [ -f ecobeehvacmode.service ];then
  cp ecobeehvacmode.service /etc/systemd/system/ecobeehvacmode.service
else
  echo "Missing ecobeehvacmode.service file"
  exit 1
fi
cp ecobeehvacmode.default /etc/default/ecobeehvacmode
mkdir /etc/ecobeehvacmode
cp ecobeehvacmode.conf /etc/ecobeehvacmode/ecobeehvacmode.conf
systemctl start ecobeehvacmode
systemctl enable ecobeehvacmode
