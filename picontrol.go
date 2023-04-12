package engine

/*
#cgo LDFLAGS:-Llib -lwiringPi

#include <stdio.h>
#include <wiringPi.h>

#define pwm0_0 1
#define pwm0_1 26
#define pwm1_0 23
#define pwm1_1 24

#define mode PWM_MODE_MS

//int lastSpeed=0;
//int lastDirection=0;

void wiringInit(){
  wiringPiSetup();
  pinMode(pwm0_1,OUTPUT);
  pinMode(pwm0_0,PWM_OUTPUT);
  pinMode(pwm1_0,PWM_OUTPUT);
  pinMode(pwm1_1,OUTPUT);
  pwmSetMode(mode);
  pwmWrite(pwm1_0,0);
  pwmWrite(pwm0_0,0);
  digitalWrite(pwm0_1,0);
  digitalWrite(pwm1_1,0);
}

void speedControl(int lastSpeed,int speed){
	if(speed >=0){
	 if(speed>16) speed=16;

	   if(lastSpeed >=0)
	   {
	     printf("1.last:%d,speed:%d\r\n",lastSpeed,speed);
	     pwmWrite(pwm0_0,64*speed);
	     pwmWrite(pwm1_0,64*speed);
	   }else{  //speed>=0 and lastSpeed<0
	          printf("2.last:%d,speed:%d\r\n",lastSpeed,speed);
			pinMode(pwm0_0,PWM_OUTPUT);
			pinMode(pwm0_1,OUTPUT);
			pinMode(pwm1_0,PWM_OUTPUT);
			pinMode(pwm1_1,OUTPUT);
			digitalWrite(pwm0_1,0);
			digitalWrite(pwm1_1,0);
			pwmSetMode(mode);
			pwmWrite(pwm0_0,64*speed);
			pwmWrite(pwm1_0,64*speed);
	    }
	 }else{  //speed<0

	  if(speed<-16) speed=-16;

	  if(lastSpeed <0) //speed<0 and lastSpeed<0
	   {
	      printf("3.last:%d,speed:%d\r\n",lastSpeed,speed);
	     pwmWrite(pwm0_1,-64*speed);
	     pwmWrite(pwm1_1,-64*speed);
	   }else{ //speed<0 and lastSpeed>=0
	      printf("4.last:%d,speed:%d\r\n",lastSpeed,speed);
			pinMode(pwm0_1,PWM_OUTPUT);
			pinMode(pwm0_0,OUTPUT);
			pinMode(pwm1_1,PWM_OUTPUT);
			pinMode(pwm1_0,OUTPUT);
			digitalWrite(pwm0_0,0);
			digitalWrite(pwm1_0,0);
			pwmSetMode(mode);
			pwmWrite(pwm0_1,-64*speed);
			pwmWrite(pwm1_1,-64*speed);
	    }
	 }//else
	//lastSpeed=speed;
	}
void directionControl(int lastSpeed,int direction){
    //int level=lastSpeed+8;. case 0:
    if(direction==0){
      speedControl(lastSpeed,lastSpeed);
      return;
      }

    int tempSpeed=0;
    printf("lastSpeed:%d,dire:%d\r\n",lastSpeed,direction);
    switch (lastSpeed){
      case -8:
      case -7:
      case -6:
      case -5:
         if(direction >=0 ){  //turn right
	  tempSpeed=lastSpeed+direction;
	  if(tempSpeed>0) tempSpeed=0;
	  pwmWrite(pwm0_0,-64*tempSpeed);
	 }else{  //turn left
	  tempSpeed=lastSpeed-direction;
	  if(tempSpeed<-16) tempSpeed=-16;
	  pwmWrite(pwm1_0,-64*tempSpeed);
	 }
       break;
      case -4:
      case -3:
      case -2:
      case -1:
         if(direction >=0 ){  //turn right
	  tempSpeed=lastSpeed-direction;
	  if(tempSpeed<-16) tempSpeed=-16;
	  pwmWrite(pwm1_0,-64*tempSpeed);
	 }else{  //turn left
	  tempSpeed=lastSpeed+direction;
	  if(tempSpeed>0) tempSpeed=0;
	  pwmWrite(pwm0_0,-64*tempSpeed);
	 }
        break;
	  case 0:
      case 1:
      case 2:
      case 3:
      case 4:
        if(direction >=0 ){  //turn right
	  tempSpeed=lastSpeed+direction;
	  if(tempSpeed>16) tempSpeed=16;
	  pwmWrite(pwm1_0,64*tempSpeed);
	 }else{  //turn left
	  tempSpeed=lastSpeed-direction;
	  if(tempSpeed>16) tempSpeed=16;
	  pwmWrite(pwm0_0,64*tempSpeed);
	 }
       break;

      case 5:
      case 6:
      case 7:
      case 8:
        if(direction >=0 ){  //turn right
	  tempSpeed=lastSpeed-direction;
	  if(tempSpeed<0) tempSpeed=0;
	  pwmWrite(pwm0_0,64*tempSpeed);
	 }else{  //turn left
	  tempSpeed=lastSpeed+direction;
	  if(tempSpeed<0) tempSpeed=0;
	  pwmWrite(pwm1_0,64*tempSpeed);
	 }
       break;
      default:break;
    }

  }
*/
import "C"

type PiControl struct {
	lastSpeed     int
	lastDirection int
}

func Init() *PiControl {
	C.wiringInit()
	return &PiControl{
		lastSpeed: 0,
	}
}
func (pi *PiControl) SpeedControl(speed int) error {
	C.speedControl(C.int(pi.lastSpeed), C.int(speed))
	pi.lastSpeed = speed
	return nil
}

func (pi *PiControl) DirectionControl(direction int) error {
	C.directionControl(C.int(pi.lastSpeed), C.int(direction))
	pi.lastDirection = direction
	return nil
}
