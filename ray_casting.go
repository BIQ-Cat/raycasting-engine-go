package main

import (
	"image/color"
	"math"
	"sync"
)

const (
	FOV             = math.Pi / 6
	RAY_DISTANCE    = 1000
	SCALE_HEIGHT    = 360
	HEIGHT_TO_COLOR = 255 / 80
)

type result struct {
	index int
	value []color.RGBA
}

type Screen struct {
	width  int
	height int
}

const MAX_MAP_SIZE int = 1920 * 1080 * 4

var mu sync.Mutex
var buffer [MAX_MAP_SIZE]uint8

func (s Screen) DeltaAngle() float64 {
	return FOV / float64(s.width)
}

type Camera struct {
	x, y     float64
	angle    float64
	height   float64
	pitch    float64
	vel      float64
	angleVel float64
}

type GameMap struct {
	Width     int      `json:"width,omitempty"`
	Height    int      `json:"height,omitempty"`
	HeightMap []int    `json:"height_map,omitempty"`
	ColorMap  [][4]int `json:"color_map,omitempty"`
}

var gameMap GameMap

var camera = Camera{
	x:        0,
	y:        0,
	angle:    math.Pi / 4,
	pitch:    -10,
	vel:      5,
	angleVel: 0.05,
	height: 	150,
}

var screen = Screen{width: 800, height: 450}

func CastRay(index int, rayAngle float64) {
	sin, cos := math.Sincos(rayAngle)
	smallestY := screen.height

	hasColorMap := gameMap.ColorMap != nil


	c := 0

	for z := 1.0; z < RAY_DISTANCE; z++ {
		y := int(z*sin + camera.y)
		if y < 0 || y >= gameMap.Height {
			continue
		}

		x := int(z*cos + camera.x)
		if x < 0 || x >= gameMap.Width {
			continue
		}

		// remove fish eye
		depth := z * math.Cos(float64(camera.angle)-rayAngle)

		heightMapIndex := y*gameMap.Width + x
		heightOnMap := gameMap.HeightMap[heightMapIndex]
		heightOnScreen := int((camera.height-float64(heightOnMap))/depth*SCALE_HEIGHT + camera.pitch)
		heightOnScreen = max(heightOnScreen, 0)



		if heightOnScreen < smallestY {

			for screenY := heightOnScreen; screenY < smallestY; screenY++ {
				grayType := (int(heightOnMap) & 0xFF) * 2

				var colorOnMap [4]int
				if hasColorMap {
					colorOnMap = gameMap.ColorMap[y*gameMap.Width+x]
				} else {
					if c % 2 == 0 {
						colorOnMap = [4]int{0, 0xFF, 0, 255}
					} else {
						colorOnMap = [4]int{0xFF, 0xFF, 0xFF, 255}
					}
					// colorOnMap = [...]int{0, 0x98, 0xDA, 255}
				}
				
				pixelIndex := screenY * screen.width + index

				buffer[pixelIndex*4] = uint8(min(0xFF, (colorOnMap[0] + grayType) / 2))
				buffer[pixelIndex*4+1] = uint8(min(0xFF, (colorOnMap[1] + grayType) / 2))
				buffer[pixelIndex*4+2] = uint8(min(0xFF, (colorOnMap[2] + grayType) / 2))
				buffer[pixelIndex*4+3] = 255
			}
			smallestY = heightOnScreen
		}
		c++
	}
}
