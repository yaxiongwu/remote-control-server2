package engine

/*
#cgo LDFLAGS:-Llib -lwiringPi

#include <stdio.h>
#include <wiringPi.h>

#define pwm0_0 1
#define pwm0_1 26
#define pwm1_0 23
#define pwm1_1 24
#define car_type_pin 4
#define mode PWM_MODE_MS

//int lastSpeed=0;
//int lastDirection=0;

int wiringInit(){
  wiringPiSetup();
  pinMode(car_type_pin,INPUT);
  pinMode(pwm0_1,OUTPUT);
  pinMode(pwm0_0,PWM_OUTPUT);
  pinMode(pwm1_0,PWM_OUTPUT);
  pinMode(pwm1_1,OUTPUT);
  pwmSetMode(mode);
  pwmWrite(pwm1_0,0);
  pwmWrite(pwm0_0,0);
  digitalWrite(pwm0_1,0);
  digitalWrite(pwm1_1,0);
  return digitalRead(car_type_pin);
}

 // 两轮差速控制小车的速度控制
void speedControl0(int lastSpeed,int speed){
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

	 // 两轮差速控制小车的方向控制
void directionControl0(int lastSpeed,int direction){
    //int level=lastSpeed+8;. case 0:
    if(direction==0){
      speedControl0(lastSpeed,lastSpeed);
      return;
      }
    direction=direction/2;//减慢速度
    int tempSpeed=0;
    printf("lastSpeed:%d,dir:%d\r\n",lastSpeed,direction);
    switch (lastSpeed){
      case -16:
      case -14:
      case -12:
      case -10:
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
      case -8:
      case -6:
      case -4:
      case -2:
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
      case 2:
      case 4:
      case 6:
      case 8:
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

      case 10:
      case 12:
      case 14:
      case 16:
     if(direction >=0 ){  //turn right
	  tempSpeed=lastSpeed-direction;
	  if(tempSpeed<0) tempSpeed=0;
	  printf("tempSpeed: %f\n", tempSpeed);
	  pwmWrite(pwm0_0,64*tempSpeed);
	 }else{  //turn left,direction<0
	  tempSpeed=lastSpeed+direction;
	  if(tempSpeed<0) tempSpeed=0;
	  pwmWrite(pwm1_0,64*tempSpeed);
	 }
       break;
      default:break;
    }

  }

//四轮阿克曼转向架控制小车的速度控制
void speedControl1(int lastSpeed,int speed){
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

	//四轮阿克曼转向架控制小车的方向控制
void directionControl1(int lastSpeed,int direction){
    //int level=lastSpeed+8;. case 0:
    if(direction==0){
      speedControl1(lastSpeed,lastSpeed);
      return;
      }
    direction=direction/2;//减慢速度
    int tempSpeed=0;
    printf("lastSpeed:%d,dir:%d\r\n",lastSpeed,direction);
    switch (lastSpeed){
      case -16:
      case -14:
      case -12:
      case -10:
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
      case -8:
      case -6:
      case -4:
      case -2:
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
      case 2:
      case 4:
      case 6:
      case 8:
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

      case 10:
      case 12:
      case 14:
      case 16:
     if(direction >=0 ){  //turn right
	  tempSpeed=lastSpeed-direction;
	  if(tempSpeed<0) tempSpeed=0;
	  printf("tempSpeed: %f\n", tempSpeed);
	  pwmWrite(pwm0_0,64*tempSpeed);
	 }else{  //turn left,direction<0
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
	carType       C.int
}

func Init() *PiControl {
	_carType := C.wiringInit()
	return &PiControl{
		lastSpeed: 0,
		carType:   _carType,
	}
}

func (pi *PiControl) SpeedControl(speed int) error {
	if pi.carType == 0 { // 两轮差速控制小车的速度控制
		C.speedControl0(C.int(pi.lastSpeed), C.int(speed))
	} else { //pi.carType == 1 四轮阿克曼转向架控制小车的速度控制
		C.speedControl1(C.int(pi.lastSpeed), C.int(speed))
	}
	pi.lastSpeed = speed
	return nil
}

// 两轮差速控制小车的方向控制
func (pi *PiControl) DirectionControl(direction int) error {
	if pi.carType == 0 { // 两轮差速控制小车的方向控制
		C.directionControl0(C.int(pi.lastSpeed), C.int(direction))
	} else { //pi.carType == 1 四轮阿克曼转向架控制小车的方向控制
		C.directionControl1(C.int(pi.lastSpeed), C.int(direction))
	}
	pi.lastDirection = direction
	return nil
}
