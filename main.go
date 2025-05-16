//go:build wasm

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"syscall/js"
)



//export getPixels
func getPixels() {
	deltaAngle := screen.DeltaAngle()

	for i := range buffer {
		buffer[i] = 0
	}
	rayAngle := camera.angle - (FOV / 2)
	for i := range screen.width {
		CastRay(i, rayAngle)
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

func loadGameMap() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		data := args[0].String()
		if err := json.Unmarshal([]byte(data), &gameMap); err != nil {
			return false
		}

		return true
	})
}

func moveCamera() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		sin, cos := math.Sincos(camera.angle)
		switch args[0].String() {
		case "w":
			camera.x += camera.vel * cos
			camera.y += camera.vel * sin
		case "s":
			camera.x -= camera.vel * cos
			camera.y -= camera.vel * sin
		case "a":
			camera.x += camera.vel * sin
			camera.y -= camera.vel * cos
		case "d":
			camera.x -= camera.vel * sin
			camera.y += camera.vel * cos
		case "ArrowUp":
			camera.pitch += camera.vel
		case "ArrowDown":
			camera.pitch -= camera.vel
		case "ArrowLeft":
			camera.angle -= camera.angleVel
		case "ArrowRight":
			camera.angle += camera.angleVel
		case " ":
			camera.height += camera.vel
		case "Shift":
			camera.height -= camera.vel
		}

		return js.ValueOf(true)
	})
}

func main() {
	noReturn := make(chan struct{})

	js.Global().Set("loadGameMap", loadGameMap())
	js.Global().Set("moveCamera", moveCamera())

	<-noReturn
	fmt.Println("here")
}
