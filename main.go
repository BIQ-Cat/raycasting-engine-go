//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"syscall/js"
)

//export loadPixels
func loadPixels(i int, showPassable bool) {
	deltaAngle := screen.DeltaAngle()

	rayAngle := camera.angle - (FOV / 2) + deltaAngle*float64(i)

	CastRay(i, rayAngle, showPassable)
}

//export clearBuffer
func clearBuffer() {
	for i := range buffer {
		buffer[i] = 0
	}
}

//export getMemoryBufferPointer
func getMemoryBufferPointer() *[MAX_MAP_SIZE]uint8 {
	return &buffer
}

//export setScreen
func setScreen(width int, height int) {
	screen.width = int(width)
	screen.height = int(height)

	if screen.width*screen.height > 1920*1080 {
		panic("max resolution is 1920 x 1080")
	}
}

func setGameMap() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		data := args[0].String()
		if err := json.Unmarshal([]byte(data), &gameMap); err != nil {
			return false
		}

		gameMap.PrepareMap()
		return true
	})
}

//export moveCamera
func moveCamera(percFb float64, percLr float64, percAngle float64, percPitch float64, percHeight float64) {
	sin, cos := math.Sincos(camera.angle)

	camera.x += percFb * camera.vel * cos
	camera.y += percFb * camera.vel * sin

	camera.x += percLr * camera.vel * sin
	camera.y -= percLr * camera.vel * cos

	camera.angle += percAngle * camera.angleVel
	camera.pitch += percPitch * camera.vel * 2

	camera.height += percHeight * camera.vel
}

func main() {
	noReturn := make(chan struct{})

	js.Global().Set("setGameMap", setGameMap())

	<-noReturn
	fmt.Println("here")
}
