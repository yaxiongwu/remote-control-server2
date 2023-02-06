package engine

import (
	"fmt"
	"time"

	"github.com/stianeikeland/go-rpio/v4"
)

type PiControl struct {
	//CarControl func(speed int, direction int)
	pinNumRunINA       int
	pinNumRunINB       int
	pinNumDirectionIN1 int
	pinNumDirectionIN2 int
	direction          int
	speed              int
}

/*
type IPiControl interface{
	CarControl func(speed int, direction int)
	}
*/
func Init(pRunINA int, pRunINB int, pDirectIN1 int, pDirectIN2 int) *PiControl {
	err := rpio.Open()
	if err != nil {
		fmt.Print(err)
		//return _,err
	}

	return &PiControl{
		pinNumRunINA:       pRunINA,
		pinNumRunINB:       pRunINB,
		pinNumDirectionIN1: pDirectIN1,
		pinNumDirectionIN2: pDirectIN2,
		direction:          0,
		speed:              0,
	}
}

func (pi *PiControl) Speed(speed int, direction bool, change <-chan int) {

}

/*
 * 速度改变时是只需要调整占空比，但是方向的Y改变时，要改变轮子的方向，此时要重新设置轮子的旋转，速度用以前的
 * */
func (pi *PiControl) DirectionControl(newDirection int) error {
	pinDirectionIN1 := rpio.Pin(pi.pinNumDirectionIN1)
	pinDirectionIN2 := rpio.Pin(pi.pinNumDirectionIN2)
	pinDirectionIN1.Output()
	pinDirectionIN2.Output()
	pinDirectionIN1.PullUp()
	pinDirectionIN2.PullUp()

	pi.direction = newDirection

	//fmt.Println(" pi.directionX:", pi.directionX)
	if pi.direction > 0 { //往左
		pinDirectionIN1.High()
		pinDirectionIN2.Low()
		//time.Sleep(time.Duration(pi.directionX) * time.Millisecond*10)
	} else if pi.direction < 0 { //往右
		pinDirectionIN1.Low()
		pinDirectionIN2.High()
		//time.Sleep(time.Duration(-pi.directionX) * time.Millisecond*10)
		//  time.Sleep(50*time.Millisecond)
	} else {
		pinDirectionIN1.Low()
		pinDirectionIN2.Low()
	}
	return nil
}
func (pi *PiControl) SpeedControl(newSpeed <-chan int) error {

	pinRunINA := rpio.Pin(pi.pinNumRunINA)
	pinRunINB := rpio.Pin(pi.pinNumRunINB)

	pinRunINA.Output() // Output mode
	pinRunINB.Output()

	pinRunINA.PullUp() //
	pinRunINB.PullUp()
	var absSpeed int

	//速度控制
	go func() {
		for {
			select {
			case s := <-newSpeed:
				pi.speed = s
				if pi.speed < 0 {
					absSpeed = -pi.speed
				} else {
					absSpeed = pi.speed
				}

				/*
					这里为了调速，需要用程序形成PWM，一共100ms，前pi.speed*10 ms运行，其余时间停止，周而复始
				*/
			case <-time.After(100*time.Millisecond - time.Duration(absSpeed*10)*time.Millisecond):
				if pi.speed > 0 { //往前跑
					pinRunINA.High()
					pinRunINB.Low()
				} else if pi.speed < 0 {
					pinRunINA.Low()
					pinRunINB.High()
				}
				time.Sleep(time.Duration(absSpeed*10) * time.Millisecond)
				//停
				pinRunINA.Low()
				pinRunINB.Low()
			} //select
		} //for
	}() //go func()
	return nil
}
