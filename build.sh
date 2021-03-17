#!/bin/sh

GOOS=linux GOARCH=arm go build -o sms_go_server .
# go build -o sms_go_server .

#sudo chmod o+rw /dev/ttyACM0

#./sim_go_server
