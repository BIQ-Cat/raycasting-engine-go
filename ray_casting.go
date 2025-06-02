package main

import (
	"math"
)

const (
	FOV             = math.Pi / 6
	RAY_DISTANCE    = 1000
	SCALE_HEIGHT    = 360
	HEIGHT_TO_COLOR = 255 / 80
)

type Screen struct {
	width  int
	height int
}

const MAX_MAP_SIZE int = 1920 * 1080 * 4

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

type Points struct {
	Flag  [][2]int `json:"flag,omitempty"`
	Relic [][2]int `json:"relic,omitempty"`
	Crit  [][2]int `json:"crit,omitempty"`
	Spawn [][2]int `json:"spawn,omitempty"`
	Slag  [][2]int `json:"slag,omitempty"`
}


type GameMap struct {
	Width          int      `json:"width,omitempty"`
	Height         int      `json:"height,omitempty"`
	HeightMap      []int    `json:"height_map,omitempty"`
	ColorMap       [][4]int `json:"color_map,omitempty"`
	PassabilityMap []bool   `json:"passability_map,omitempty"`
	Points                  `json:"points,omitempty"`
}

func (g *GameMap) IsPassable(x int, y int) bool {
	passabilityMapWidth := (g.Width - 1) / 2
	passabilityMapHeight := (g.Height - 1) / 2

	unusedSurfaceX := g.Width / 4
	if x < unusedSurfaceX || x >= (g.Width-unusedSurfaceX) {
		return false
	}

	unusedSurfaceY := g.Height / 4
	if y < unusedSurfaceY || y >= (g.Height-unusedSurfaceY) {
		return false
	}

	passX := x - unusedSurfaceX
	passY := passabilityMapHeight - (y - unusedSurfaceY)

	pass1 := true
	pass2 := true
	pass3 := true
	pass4 := true

	if passX != 0 && passY != 0 {
		pass1 = g.PassabilityMap[(passY-1)*passabilityMapWidth+(passX-1)]
	}

	if passX != 0 && passY != passabilityMapHeight {
		pass2 = g.PassabilityMap[(passY)*passabilityMapWidth+(passX-1)]
	}

	if passX != passabilityMapWidth && passY != 0 {
		pass3 = g.PassabilityMap[(passY-1)*passabilityMapWidth+(passX)]
	}

	if passX != passabilityMapWidth && passY != passabilityMapHeight {
		pass4 = g.PassabilityMap[(passY)*passabilityMapWidth+(passX)]
	}

	return pass1 && pass2 && pass3 && pass4
}

func (g *GameMap) PrepareMap() {
	centerX := g.Width / 2
	centerY := g.Height / 2
	if g.Flag != nil {
		for _, flagCoords := range g.Flag {
			g.drawEntity(flagCoords, centerX, centerY, FLAG_COLOR_MAP, FLAG_HEIGHT_MAP)
		}
	}


	if g.Crit != nil {
		for _, critCoords := range g.Crit {
			g.drawEntity(critCoords, centerX, centerY, CRIT_COLOR_MAP, CRIT_HEIGHT_MAP)
		}
	}

	if g.Relic != nil {
		for _, relicCoords := range g.Relic {
			g.drawEntity(relicCoords, centerX, centerY, RELIC_COLOR_MAP, RELIC_HEIGHT_MAP)
		}
	}

	if g.Spawn != nil {
		for _, spawnCoords := range g.Spawn {
			g.drawEntity(spawnCoords, centerX, centerY, SPAWN_COLOR_MAP, SPAWN_HEIGHT_MAP)
		}
	}

	if g.Slag != nil {
		for _, slagCoords := range g.Slag {
			g.drawEntity(slagCoords, centerX, centerY, SLAG_COLOR_MAP, SLAG_HEIGHT_MAP)
		}
	}
}

func (g *GameMap) drawEntity(entityCoords [2]int, centerX int, centerY int, entityColorMap [][][4]int, entityHeightMap [][]int) {
	startX := centerX + entityCoords[0] / 2 - len(entityHeightMap[0]) / 2 - 1
	startY := centerY + entityCoords[1] / 2 - len(entityHeightMap) / 2 - 1

	for y := range(len(entityHeightMap)) {
		for x := range(len(entityHeightMap[y])) {
			g.HeightMap[(y + startY) * g.Width + (x + startX)] += entityHeightMap[y][x]

			color := entityColorMap[y][x]
			if color[3] != 0 {
				g.ColorMap[(g.Height - y - startY - 1) * g.Width + (x + startX)] = color
			}
		}
	}
}

var gameMap GameMap

var camera = Camera{
	x:        0,
	y:        0,
	angle:    math.Pi / 4,
	pitch:    -10,
	vel:      2,
	angleVel: 0.02,
	height:   150,
}

var screen = Screen{width: 800, height: 450}

func CastRay(index int, rayAngle float64, showPassable bool) {
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

		mapIndex := (gameMap.Height-y-1)*gameMap.Width + x

		heightOnMap := gameMap.HeightMap[mapIndex]
		heightOnScreen := int((camera.height-float64(heightOnMap))/depth*SCALE_HEIGHT + camera.pitch)
		heightOnScreen = max(heightOnScreen, 0)

		if heightOnScreen < smallestY {

			for screenY := heightOnScreen; screenY < smallestY; screenY++ {
				grayType := (int(heightOnMap) & 0xFF) * 2

				var colorOnMap [4]int
				if hasColorMap {
					colorOnMap = gameMap.ColorMap[y*gameMap.Width+x]
				} else {
					if c%2 == 0 {
						colorOnMap = [4]int{0, 0xFF, 0, 255}
					} else {
						colorOnMap = [4]int{0xFF, 0xFF, 0xFF, 255}
					}
				}

				if showPassable && !gameMap.IsPassable(x, y) {
					colorOnMap = [4]int{0xFF, 0, 0, 255}
				}

				pixelIndex := screenY*screen.width + index

				buffer[pixelIndex*4] = uint8(min(0xFF, (colorOnMap[0]+grayType)/2))
				buffer[pixelIndex*4+1] = uint8(min(0xFF, (colorOnMap[1]+grayType)/2))
				buffer[pixelIndex*4+2] = uint8(min(0xFF, (colorOnMap[2]+grayType)/2))
				buffer[pixelIndex*4+3] = 255
			}
			smallestY = heightOnScreen
		}
		c++
	}
}
