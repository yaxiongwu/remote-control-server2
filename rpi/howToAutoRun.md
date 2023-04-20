在自动重启过程中，出现了bug，花费了很长时间才解决。
问题：调用硬件pwm需要sudo权限，在接了hdmi的图形界面中使用Terminal运行没有问题，但是自启动或者没有接hdmi远程启动时也有问题，
   两种情况报同一个错误：
   AL lib: (EE) ALCcaptureAlsa_open: Could not open capture device 'default': No such file or directory。
   
   最开始一直纠结于su - root ,su root,su - pi,su pi等各种登录方式的区别中，因为表面上看是登录不同用户造成的。
   实际的bug应该是接了hdmi屏幕的pi tty7用户初始化时检测到了hdmi音频，使用cat /proc/asound/cards命令查看有三个音频设备，
   0：HDMI
   1:headphones
   2:USB Pnp Sound Device
   但是如果从网络启动，或者从自动运行时，从root到pi用户，这个新的用户pi tty1，没有使用音频的权限或者环境变量，
   通过两者的env命令输出的环境变量比较，pi tty7(hdmi图形界面登录)的env比pi tty1的env多出很多参数，其中加入下列两个就能成功运行
   export XAUTHORITY=/home/pi/.Xauthority
   export DISPLAY=:0.0
   
   如果要小车要行走，没有接hdmi时，只有两个音频设备：
   0:headphones
   1:USB Pnp Sound Device
  
   而在调试程序的时候，连接了HDMI，开始就用sudo amixer cset numid=3 1，使得root里设置了播放设备为1,有HDMI时，1为headphones，
  没有HDMI，1变成了USB 录音设备，无法播放声音，导致无法打开的错误。
   所以在没有连接hdmi时，应该设置sudo amixer cset numid=3 0。
   命令speaker-test  -t3 测试声音输出
  

1.go build生成可执行文件
2.建一个sh文件,内容如下
    #!/bin/bash
    cd /home/pi/remote-control-client-go2-0927/rpi
    source ./env
    sudo chmod +x main
    sudo nohup ./main &    
3.chmod +x *.sh  
  chmod +x /etc/rc.local 
4.在/etc/rc.local文件的
    exit 0前添加：
    sleep 30
    su - pi -c "bash /home/pi/remote-control-client-go2-0927/rpi/autoRunWebrtc.sh"
