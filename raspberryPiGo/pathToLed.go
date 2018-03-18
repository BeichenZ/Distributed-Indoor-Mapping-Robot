package main

import (
	shared "./sharedsis"
	"github.com/stianeikeland/go-rpio"
	"fmt"
	"os"
	"time"
)

var (
	pinNorth = rpio.Pin(9)
	pinSouth = rpio.Pin(11)
	pinEast = rpio.Pin(25)
	pinWest = rpio.Pin(8)

	pinFreeSpaceButton = rpio.Pin(5)
	pinWallButton = rpio.Pin(6)
	pinRightBumperButton = rpio.Pin(12)


)
func main () {
	// Pin setup
	if err := rpio.Open(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Unmap gpio memory when done
	defer rpio.Close()

	// Set pins to output mode
	pinNorth.Output()
	pinSouth.Output()
	pinEast.Output()
	pinWest.Output()

	//// TESTING //
	//// Toggle pin 20 times
	//for x := 0; x < 20; x++ {
	//	pinNorth.Toggle()
	//	time.Sleep(time.Second / 5)
	//}
	//// TESTING //

	path := [...]shared.PointStruct{shared.WEST, shared.WEST, shared.NORTH}

	for _,dir := range path{
		switch dir {
			case shared.NORTH: {
				pinNorth.High()
				time.Sleep(2*time.Second)
				pinNorth.Low()
				break;
			}
			case shared.SOUTH: {
				pinSouth.High()
				time.Sleep(2*time.Second)
				pinSouth.Low()
				break;
			}
			case shared.EAST: {
				pinEast.High()
				time.Sleep(2*time.Second)
				pinEast.Low()
				break;
			}
			case shared.WEST: {
				pinWest.High()
				time.Sleep(2*time.Second)
				pinWest.Low()
				break;
			}

		}
	}

}

