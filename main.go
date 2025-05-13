//go:build wasm

package main

import (
	"encoding/json"
	"fmt"
	"math"
	"syscall/js"
)

func getPixels(gameMap *GameMap) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		data := make(chan result, screen.height)
		pixels := make([]any, screen.height*screen.width*4)

		deltaAngle := screen.DeltaAngle()

		rayAngle := camera.angle - (FOV / 2)
		for i := range screen.width {
			go CastRay(i, gameMap, rayAngle, data)
			rayAngle += deltaAngle
		}

		for range screen.width {
			line := <-data
			for i, el := range line.value {
				index := i*screen.width + line.index
				pixels[index*4] = el.R
				pixels[index*4+1] = el.G
				pixels[index*4+2] = el.B
				pixels[index*4+3] = el.A
			}
		}

		return js.ValueOf(pixels)
	})

}

func setScreen() js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		screen.width = args[0].Int()
		screen.height = args[1].Int()

		return js.Undefined()
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
			camera.x += camera.vel * cos
			camera.y -= camera.vel * sin
		case "d":
			camera.x -= camera.vel * cos
			camera.y += camera.vel * sin
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

	data := js.Global().Get("gameMap")

	var gameMap GameMap

	err := json.Unmarshal([]byte(data.String()), &gameMap)

	if err != nil {
		panic(err)
	}

	js.Global().Set("getPixels", getPixels(&gameMap))
	js.Global().Set("setScreen", setScreen())
	js.Global().Set("moveCamera", moveCamera())

	<-noReturn
	fmt.Println("here")
}
