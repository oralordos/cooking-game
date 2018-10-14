package main

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten"
)

const (
	speed       = 100.0 / 60.0
	maxWorkDist = 20
)

type input struct {
	moveHorizontal float64
	moveVertical   float64
	interact       bool
	startWorking   bool
}

type controller interface {
	getInput() input
}

type keyboardController struct {
	keyMap    map[ebiten.Key]ebiten.Key
	lastSpace bool
	lastE     bool
}

func newKeyboardController() *keyboardController {
	return &keyboardController{
		keyMap: map[ebiten.Key]ebiten.Key{
			ebiten.KeyA:     ebiten.KeyA,
			ebiten.KeyD:     ebiten.KeyD,
			ebiten.KeyW:     ebiten.KeyW,
			ebiten.KeyS:     ebiten.KeyS,
			ebiten.KeySpace: ebiten.KeyE,
			ebiten.KeyE:     ebiten.KeyQ,
		},
		lastSpace: false,
		lastE:     false,
	}
}

func newKeyboardControllerKeyMap(keyMap map[ebiten.Key]ebiten.Key) *keyboardController {
	return &keyboardController{
		keyMap:    keyMap,
		lastSpace: false,
		lastE:     false,
	}
}

func (kc *keyboardController) getInput() input {
	hori := 0.0
	vert := 0.0

	if ebiten.IsKeyPressed(kc.keyMap[ebiten.KeyA]) {
		hori -= 1
	}

	if ebiten.IsKeyPressed(kc.keyMap[ebiten.KeyD]) {
		hori += 1
	}

	if ebiten.IsKeyPressed(kc.keyMap[ebiten.KeyW]) {
		vert -= 1
	}

	if ebiten.IsKeyPressed(kc.keyMap[ebiten.KeyS]) {
		vert += 1
	}

	space := ebiten.IsKeyPressed(kc.keyMap[ebiten.KeySpace])
	inter := !kc.lastSpace && space
	kc.lastSpace = space

	e := ebiten.IsKeyPressed(kc.keyMap[ebiten.KeyE])
	work := !kc.lastE && e
	kc.lastE = e

	return input{
		moveHorizontal: hori,
		moveVertical:   vert,
		interact:       inter,
		startWorking:   work,
	}
}

type gamepadController struct {
	controllerIndex int
	lastA           bool
	lastX           bool
}

func newGamepadController(index int) *gamepadController {
	return &gamepadController{
		controllerIndex: index,
	}
}

func (gc *gamepadController) getInput() input {
	controllerIDs := ebiten.GamepadIDs()
	if gc.controllerIndex >= len(controllerIDs) {
		return input{}
	}

	cID := controllerIDs[gc.controllerIndex]

	a := ebiten.IsGamepadButtonPressed(cID, ebiten.GamepadButton0)
	inter := !gc.lastA && a
	gc.lastA = a

	x := ebiten.IsGamepadButtonPressed(cID, ebiten.GamepadButton2)
	work := !gc.lastX && x
	gc.lastX = x

	return input{
		moveHorizontal: ebiten.GamepadAxis(cID, 0),
		moveVertical:   -ebiten.GamepadAxis(cID, 1),
		interact:       inter,
		startWorking:   work,
	}
}

type player struct {
	x       float64
	y       float64
	img     *image.Rectangle
	control controller
	holding item
	working bool
}

func newPlayer(x, y float64, img *image.Rectangle, control controller) *player {
	return &player{
		x:       x,
		y:       y,
		img:     img,
		control: control,
		holding: nil,
		working: false,
	}
}

func (p *player) draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{
		SourceRect: p.img,
	}
	op.GeoM.Translate(p.x, p.y)
	op.GeoM.Scale(2, 2)
	screen.DrawImage(spriteSheet, op)

	if p.holding != nil {
		p.holding.draw(screen, p.x, p.y)
	}
}

func (p *player) work() {
	input := p.control.getInput()

	hori := input.moveHorizontal / (math.Abs(input.moveHorizontal) + math.Abs(input.moveVertical))
	vert := input.moveVertical / (math.Abs(input.moveHorizontal) + math.Abs(input.moveVertical))
	if input.moveHorizontal == 0 && input.moveVertical == 0 {
		hori = 0
		vert = 0
	} else {
		p.working = false
	}

	p.x += speed * hori
	p.y += speed * vert

	for _, t := range tiles {
		tx, ty := t.getPosition()
		px, py := p.x+8, p.y+8

		dist := getDistance(px, py, tx, ty)
		if dist > 8+math.Sqrt(8*8+8*8) {
			continue
		}

		if dist < 16 {
			p.x -= speed * hori
			p.y -= speed * vert
		}

		if p.x >= tx-8 && p.x < tx+8 && p.y >= ty-8 && p.y < ty+8 {
			p.x -= speed * hori
			p.y -= speed * vert
		}
	}

	for _, otherP := range players {
		if otherP == p {
			continue
		}

		if getDistance(p.x, p.y, otherP.x, otherP.y) < 16 {
			p.x -= speed * hori
			p.y -= speed * vert
		}
	}

	if input.interact {
		t := getNearbyTile(p.x+8, p.y+8, maxWorkDist)
		if t != nil {
			if p.holding == nil {
				p.holding = t.pickUp()
			} else {
				if t.drop(p.holding) {
					p.holding = nil
				}
			}
		}
	}

	if input.startWorking {
		p.working = true
	}

	if p.working {
		t := getNearbyTile(p.x+8, p.y+8, maxWorkDist)
		if t != nil {
			t.work()
		}
	}
}

func getNearbyTile(x, y, maxDist float64) tile {
	closestDist := maxDist
	var closest tile = nil

	for _, t := range tiles {
		tx, ty := t.getPosition()
		dist := getDistance(x, y, tx, ty)
		if dist < closestDist {
			closestDist = dist
			closest = t
		}
	}

	return closest
}

func getDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}
