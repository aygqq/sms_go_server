#!/bin/sh

if [[ -f ./sms_go_server && -f ./sms_go_server_upd ]]; then
    mv sms_go_server         sms_go_server_running
    mv sms_go_server_upd     sms_go_server
    rm sms_go_server_running
    
    chmod +x sms_go_server

    echo "Files chandeg successful. Rebooting."
    shutdown -r now
else
    echo "Error: Some of the files not exists"
fi
