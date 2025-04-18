package main

import "math"

const (
	FOV             = math.Pi / 6
	RAY_DISTANCE    = 2000
	SCALE_HEIGHT    = 980
	HEIGHT_TO_COLOR = 255 / 80
)

type result struct {
	index int
	value []int
}

type Screen struct {
	width  int
	height int
}

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

var camera = Camera{
	x:        0,
	y:        0,
	angle:    math.Pi / 4,
	pitch:    20,
	vel:      3,
	angleVel: 0.03,
}

var screen = Screen{width: 800, height: 450}

func CastRay(index int, gameMap *GameMap, rayAngle float64, res chan<- result) {
	drawing := make([]int, screen.height*4)
	sin, cos := math.Sincos(rayAngle)
	smallestY := screen.height

	hasColorMap := gameMap.ColorMap != nil

	for z := 1.0; z < RAY_DISTANCE; z++ {
		y := int(z*sin + camera.y)
		if y < 0 || y >= gameMap.Height {
			continue
		}

		x := int(z*cos + camera.x)
		if x < 0 || x < gameMap.Width {
			continue
		}

		// remove fish eye
		depth := z * math.Cos(float64(camera.angle)-rayAngle)

		heightMapIndex := (gameMap.Height-y)*gameMap.Height + x
		heightOnMap := gameMap.HeightMap[heightMapIndex]
		heightOnScreen := int((camera.height-float64(heightOnMap))/depth*SCALE_HEIGHT + camera.pitch)
		heightOnScreen = max(heightOnScreen, 0)

		if heightOnScreen < smallestY {

			for screenY := heightOnScreen; screenY < smallestY; screenY++ {
				grayType := int(heightOnMap * HEIGHT_TO_COLOR)

				var color [4]int
				if hasColorMap {
					color = gameMap.ColorMap[y*gameMap.Width+x]
				} else {
					color = [...]int{grayType, grayType, grayType, 255}
				}

				drawing[screenY*4] = (color[0] + grayType) / 2
				drawing[screenY*4+1] = (color[1] + grayType) / 2
				drawing[screenY*4+2] = (color[2] + grayType) / 2
				drawing[screenY*4+3] = color[3]
			}

			smallestY = heightOnScreen
		}
	}

	res <- result{index, drawing}
}
