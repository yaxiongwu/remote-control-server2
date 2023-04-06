#!/bin/bash
cd /home/pi/webrtc/remote-control-client-go2/rpi
#source ./env
sudo amixer cset numid=3 0
sudo chmod +x main
sudo nohup /home/pi/webrtc/remote-control-client-go2/rpi/main &
