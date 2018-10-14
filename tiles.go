package main

import (
	"image"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten"
)

func drawProgress(screen *ebiten.Image, x, y, progressPercent float64, c color.Color) {
	progress, _ := ebiten.NewImage(16, 3, ebiten.FilterDefault)
	defer progress.Dispose()
	progress.Fill(c)

	progressWidth := int(14 * progressPercent)
	if progressWidth <= 0 {
		progressWidth = 1
	}
	bar, _ := ebiten.NewImage(progressWidth, 1, ebiten.FilterDefault)
	defer bar.Dispose()
	bar.Fill(color.White)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	op.GeoM.Scale(2, 2)
	screen.DrawImage(progress, op)

	op = &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x+1, y+1)
	op.GeoM.Scale(2, 2)
	screen.DrawImage(bar, op)
}

type baseTile struct {
	x   float64
	y   float64
	img *image.Rectangle
}

func (bt *baseTile) getPosition() (float64, float64) {
	return bt.x + 8, bt.y + 8
}

func (bt *baseTile) draw(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{
		SourceRect: bt.img,
	}
	op.GeoM.Translate(bt.x, bt.y)
	op.GeoM.Scale(2, 2)
	screen.DrawImage(spriteSheet, op)
}

func (bt *baseTile) work() {}

type wall struct {
	baseTile
}

func newWall(x, y float64) *wall {
	return &wall{
		baseTile: baseTile{
			x:   x,
			y:   y,
			img: &wallImage,
		},
	}
}

func (w *wall) pickUp() item {
	return nil
}

func (w *wall) drop(i item) bool {
	return false
}

type counter struct {
	baseTile
	contents item
}

func newCounter(x, y float64, contents item) *counter {
	return &counter{
		baseTile: baseTile{
			x:   x,
			y:   y,
			img: &counterImage,
		},
		contents: contents,
	}
}

func (c *counter) draw(screen *ebiten.Image) {
	c.baseTile.draw(screen)
	if c.contents != nil {
		c.contents.draw(screen, c.baseTile.x, c.baseTile.y)
	}
}

func (c *counter) pickUp() item {
	if c.contents != nil {
		i := c.contents
		c.contents = nil
		return i
	}

	return nil
}

func (c *counter) drop(i item) bool {
	if c.contents != nil {
		return c.contents.combine(i)
	}

	c.contents = i
	return true
}

type tomatoSpawner struct {
	baseTile
}

func newTomatoSpawner(x, y float64) *tomatoSpawner {
	return &tomatoSpawner{
		baseTile: baseTile{
			x:   x,
			y:   y,
			img: &tomatoSpawnerImage,
		},
	}
}

func (ts *tomatoSpawner) pickUp() item {
	return newTomato()
}

func (ts *tomatoSpawner) drop(i item) bool {
	return false
}

type lettuceSpawner struct {
	baseTile
}

func newLettuceSpawner(x, y float64) *lettuceSpawner {
	return &lettuceSpawner{
		baseTile: baseTile{
			x:   x,
			y:   y,
			img: &lettuceSpawnerImage,
		},
	}
}

func (ts *lettuceSpawner) pickUp() item {
	return newLettuce()
}

func (ts *lettuceSpawner) drop(i item) bool {
	return false
}

type exit struct {
	baseTile
	numDirty int
}

func newExit(x, y float64) *exit {
	return &exit{
		baseTile: baseTile{
			x:   x,
			y:   y,
			img: &exitImage,
		},
		numDirty: 0,
	}
}

func (e *exit) draw(screen *ebiten.Image) {
	e.baseTile.draw(screen)

	if e.numDirty > 0 {
		newDirtyPlate().draw(screen, e.x+8, e.y+8)
	}
}

func (e *exit) pickUp() item {
	if e.numDirty > 0 {
		e.numDirty--
		return newDirtyPlate()
	}
	return nil
}

func (e *exit) drop(i item) bool {
	if cp, ok := i.(*cleanPlate); ok {
		plateContents := make([]item, 0, len(cp.contents))
		for _, v := range cp.contents {
			plateContents = append(plateContents, v)
		}
		sort.Slice(plateContents, func(i, j int) bool {
			return plateContents[i].name() < plateContents[j].name()
		})
		for oi, order := range orders {
			if len(order) != len(plateContents) {
				continue
			}

			matching := true
			for ii := 0; ii < len(order); ii++ {
				if order[ii].name() != plateContents[ii].name() {
					matching = false
					break
				}
			}
			if matching {
				orders = append(orders[:oi], orders[oi+1:]...)
				e.numDirty++
				return true
			}
		}
	}
	return false
}

type trash struct {
	baseTile
}

func newTrash(x, y float64) *trash {
	return &trash{
		baseTile: baseTile{
			x:   x,
			y:   y,
			img: &trashImage,
		},
	}
}

func (t *trash) pickUp() item {
	return nil
}

func (t *trash) drop(i item) bool {
	if p, ok := i.(*cleanPlate); ok {
		p.contents = map[string]item{}
		return false
	}
	if i.name() == "dirtyPlate" {
		return false
	}

	return true
}

type choppingBoard struct {
	baseTile
	contents item
	progress float64
}

func newChoppingBoard(x, y, progress float64, contents item) *choppingBoard {
	return &choppingBoard{
		baseTile: baseTile{
			x:   x,
			y:   y,
			img: &choppingBoardImage,
		},
		contents: contents,
		progress: progress,
	}
}

func (cb *choppingBoard) draw(screen *ebiten.Image) {
	cb.baseTile.draw(screen)

	if cb.contents != nil {
		cb.contents.draw(screen, cb.x+8, cb.y+8)
	}

	if cb.progress > 0 {
		drawProgress(screen, cb.x, cb.y, cb.progress, color.RGBA{0, 255, 0, 255})
	}
}

func (cb *choppingBoard) pickUp() item {
	if cb.contents != nil {
		i := cb.contents
		cb.contents = nil
		cb.progress = 0
		return i
	}

	return nil
}

func (cb *choppingBoard) drop(i item) bool {
	if !i.canCut() {
		return false
	}

	if cb.contents != nil {
		return cb.contents.combine(i)
	}

	cb.contents = i
	return true
}

func (cb *choppingBoard) work() {
	if cb.contents == nil || !cb.contents.canCut() {
		return
	}

	cb.progress += 1.0 / 3.0 / 60.0 // 3 seconds to complete
	if cb.progress >= 1 {
		cb.progress = 0
		cb.contents = cb.contents.completeCut()
	}
}

type sink struct {
	baseTile
	contents item
	progress float64
}

func newSink(x, y, progress float64, contents item) *sink {
	return &sink{
		baseTile: baseTile{
			x:   x,
			y:   y,
			img: &sinkImage,
		},
		contents: contents,
		progress: progress,
	}
}

func (s *sink) draw(screen *ebiten.Image) {
	s.baseTile.draw(screen)

	if s.contents != nil {
		s.contents.draw(screen, s.x+8, s.y+8)
	}

	if s.progress > 0 {
		drawProgress(screen, s.x, s.y, s.progress, color.RGBA{0, 0, 255, 255})
	}
}

func (s *sink) pickUp() item {
	if s.contents != nil {
		i := s.contents
		s.contents = nil
		s.progress = 0
		return i
	}

	return nil
}

func (s *sink) drop(i item) bool {
	if !i.canWash() {
		return false
	}

	if s.contents != nil {
		return s.contents.combine(i)
	}

	s.contents = i
	return true
}

func (s *sink) work() {
	if s.contents == nil || !s.contents.canWash() {
		return
	}

	s.progress += 1.0 / 3.0 / 60.0 // 3 seconds to complete
	if s.progress >= 1 {
		s.progress = 0
		s.contents = s.contents.completeWash()
	}
}
