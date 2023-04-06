在自动重启过程中，出现了bug，花费了很长时间才解决。
问题：调用硬件pwm需要sudo权限，但是在Terminal中运行程序不会有问题，自动启动时，会报错退出：
   AL lib: (EE) ALCcaptureAlsa_open: Could not open capture device 'default': No such file or directory。
相关知识：
  在用户切换是su pi或者su root，只是切换运行权限，调用和运行的环境还是原来的，可以用env命令测试。
  要想完全切换，使用su  - pi，或者su - root。
  自动启动尝试有两种方法可行，一是在/etc/rc.local添加启动代码，二是在/etc/init.d中新增服务，相关命令：sduo systemctl deamon-reloal && sudo service ** stop && sudo service start.
  这两种方式启动时使用的都是完全root用户环境，不管如何切换到pi，跟直接使用Terminal终端的env命令输出结果不一致，可以在Terminal中使用su - root，再su - pi，然后env，与直接打开Terminal，env结果对比。
  这里可能涉及到交互式、登录与非登录的区别问题，但是以bash -il 命令也没能解决问题。
  最终在直接打开Terminal中env得到的参数保存到env文件中，去掉su - root && su - pi &&env中出现的，source ./env，使得环境跟Terminal中的当前实际用户环境一致，可以解决这个问题。
  不一样的地方那个很多，但是最终只有这两项起作用：
  export XAUTHORITY=/home/pi/.Xauthority
  export DISPLAY=:0.0
  具体功能待分析。
  但为什么会报 ”AL lib: (EE) ALCcaptureAlsa_open“ 错误值得研究，暂时没有时间深究。另外如何启动后成为当前用户的环境也值得研究。
  问题解决：当使用连接HDMI时，音频有三个,使用cat /proc/asound/cards命令查看，
   0：HDMI
   1:headphones
   2:USB Pnp Sound Device
   但是如果没有使用HDMI时，音频只有两个：
   0:headphones
   1:USB Pnp Sound Device
   而在调试程序的时候，连接了HDMI，开始就用sudo amixer cset numid=3 1，使得root里设置了播放设备为1,有HDMI时，1为headphones，
   但是在断开HDMI，自动运行时，没有了HDMI，1变成了USB 录音设备，无法播放声音，导致无法打开的错误。
   所以应该设置sudo amixer cset numid=3 0
  

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
