#!/bin/sh

cd /usr/local/projects/OpenAMP_TTY_CM4
./fw_cortex_m4.sh start

stty -onlcr -echo -F /dev/ttyRPMSG0

cd /home/root/hello
./sms_mp1