#!/usr/bin/expect -f
set timeout 20
spawn su root
expect "Password:"
send "2wsx4rfv^YHN*IK<\r"
expect "root@raspberrypi:/home/pi/webrtc/remote-control-client-go2/rpi#"
#spawn cd /home/pi/webrtc/remote-control-client-go2/rpi
send "nohup ./main &\r"
interact
