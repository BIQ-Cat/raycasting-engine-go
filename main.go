//go:build wasm

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"syscall/js"
)

//export loadPixels
func loadPixels(showPassable bool) {
	deltaAngle := screen.DeltaAngle()

	for i := range buffer {
		buffer[i] = 0
	}
	rayAngle := camera.angle - (FOV / 2)
	for i := range screen.width {
		CastRay(i, rayAngle, showPassable)
		rayAngle += deltaAngle
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
func moveCamera(perc_fb float64, perc_lr float64, perc_angle float64, perc_pitch float64, perc_height float64) {
	sin, cos := math.Sincos(camera.angle)

	camera.x += perc_fb * camera.vel * cos
	camera.y += perc_fb * camera.vel * sin

	camera.x += perc_lr * camera.vel * sin
	camera.y -= perc_lr * camera.vel * cos

	camera.angle += perc_angle * camera.angleVel
	camera.pitch += perc_pitch * camera.vel * 2

	camera.height += perc_height * camera.vel
}

func main() {
	noReturn := make(chan struct{})

	js.Global().Set("setGameMap", setGameMap())

	<-noReturn
	fmt.Println("here")
}
