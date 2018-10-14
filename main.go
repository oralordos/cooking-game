package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"math/rand"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

const (
	windowTitle  = "Cooking"
	screenWidth  = 640
	screenHeight = 480
	scale        = 1.5
)

var (
	spriteSheet         *ebiten.Image
	wallImage           = image.Rect(0, 0, 16, 16)
	counterImage        = image.Rect(16, 0, 32, 16)
	tomatoSpawnerImage  = image.Rect(32, 0, 48, 16)
	lettuceSpawnerImage = image.Rect(48, 0, 64, 16)
	exitImage           = image.Rect(0, 16, 16, 32)
	trashImage          = image.Rect(16, 16, 32, 32)
	choppingBoardImage  = image.Rect(32, 16, 48, 32)
	sinkImage           = image.Rect(48, 16, 64, 32)
	player1Image        = image.Rect(0, 32, 16, 48)
	player2Image        = image.Rect(16, 32, 32, 48)
	cleanPlateImage     = image.Rect(32, 32, 40, 40)
	dirtyPlateImage     = image.Rect(32, 40, 40, 48)
	tomatoImage         = image.Rect(40, 32, 48, 40)
	choppedTomatoImage  = image.Rect(40, 40, 48, 48)
	lettuceImage        = image.Rect(48, 32, 56, 40)
	choppedLettuceImage = image.Rect(48, 40, 56, 48)
)

func init() {
	rawImgData := MustAsset("assets/tiles.png")
	img, _, err := image.Decode(bytes.NewReader(rawImgData))
	if err != nil {
		panic(err)
	}

	rgba := image.NewRGBA(img.Bounds())
	for y := img.Bounds().Min.Y; y <= img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x <= img.Bounds().Max.X; x++ {
			c := img.At(x, y)
			r, g, b, _ := c.RGBA()
			if r == 0 && g == 0 && b == 0 {
				rgba.SetRGBA(x, y, color.RGBA{
					R: 0,
					G: 0,
					B: 0,
					A: 0,
				})
			} else {
				rgba.Set(x, y, c)
			}
		}
	}

	spriteSheet, _ = ebiten.NewImageFromImage(rgba, ebiten.FilterDefault)
}

type item interface {
	draw(screen *ebiten.Image, x, y float64)
	combine(other item) bool
	canCombine() bool
	name() string
	canCut() bool
	completeCut() item
	canWash() bool
	completeWash() item
}

type tile interface {
	getPosition() (float64, float64)
	draw(screen *ebiten.Image)
	pickUp() item
	drop(i item) bool
	work()
}

var tiles = []tile{
	newWall(10, 10),
	newTrash(26, 10),
	newChoppingBoard(42, 10, 0.25, newLettuce()),
	newTomatoSpawner(58, 10),
	newLettuceSpawner(74, 10),
	newExit(90, 10),
	newSink(106, 10, 0.25, newDirtyPlate()),
	newCounter(10, 26, newCleanPlate([]item{newChoppedLettuce(), newChoppedTomato()})),
	newCounter(26, 26, newLettuce()),
	newCounter(42, 26, newChoppedLettuce()),
	newCounter(58, 26, newTomato()),
	newCounter(74, 26, newChoppedTomato()),
	newCounter(90, 26, newDirtyPlate()),
	newCounter(106, 26, nil),
}

var players = []*player{
	newPlayer(100, 100, &player1Image, newKeyboardController()),
	newPlayer(150, 150, &player2Image, newKeyboardControllerKeyMap(map[ebiten.Key]ebiten.Key{
		ebiten.KeyA:     ebiten.KeyJ,
		ebiten.KeyD:     ebiten.KeyL,
		ebiten.KeyW:     ebiten.KeyI,
		ebiten.KeyS:     ebiten.KeyK,
		ebiten.KeySpace: ebiten.KeyO,
		ebiten.KeyE:     ebiten.KeyU,
	})),
	newPlayer(100, 150, &player2Image, newGamepadController(0)),
}

var orders = [][]item{
	{
		newChoppedLettuce(),
		newChoppedTomato(),
	},
}

var possibleOrders = [][]item{
	{
		newChoppedLettuce(),
	},
	{
		newChoppedLettuce(),
		newChoppedTomato(),
	},
}

var gamepadIDs = map[int]struct{}{}

var startup = true

func update(screen *ebiten.Image) error {
	if startup {
		ebiten.SetCursorVisible(false)
		startup = false
	}

	for _, p := range players {
		p.work()
	}

	if len(orders) < 5 {
		if rand.Intn(60*10) == 0 { // Approximately one per 10 seconds
			newOrder := possibleOrders[rand.Intn(len(possibleOrders))]
			orders = append(orders, newOrder)
		}
	}

	if ebiten.IsDrawingSkipped() {
		return nil
	}

	for _, t := range tiles {
		t.draw(screen)
	}

	for _, p := range players {
		p.draw(screen)
	}

	_, h := screen.Size()
	for oi, o := range orders {
		for ii, i := range o {
			ebitenutil.DebugPrintAt(screen, i.name(), 0, h-20-oi*40-ii*10)
		}
	}

	ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %0.2f\nTPS: %0.2f", ebiten.CurrentFPS(), ebiten.CurrentTPS()))

	return nil
}

func main() {
	defer spriteSheet.Dispose()
	err := ebiten.Run(update, screenWidth, screenHeight, scale, windowTitle)
	if err != nil {
		panic(err)
	}
}
