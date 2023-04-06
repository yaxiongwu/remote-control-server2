#!/bin/bash
cd /home/pi/webrtc/remote-control-client-go2/rpi
source ./env
sudo chmod +x main
sudo nohup /home/pi/webrtc/remote-control-client-go2/rpi/main &
