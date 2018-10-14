package main

import (
	"image"
	"sort"

	"github.com/hajimehoshi/ebiten"
)

type baseItem struct {
	img *image.Rectangle
}

func (bi *baseItem) draw(screen *ebiten.Image, x, y float64) {
	op := &ebiten.DrawImageOptions{
		SourceRect: bi.img,
	}
	op.GeoM.Translate(x, y)
	op.GeoM.Scale(2, 2)
	screen.DrawImage(spriteSheet, op)
}

func (bi *baseItem) combine(other item) bool {
	return false
}

func (bi *baseItem) canCut() bool {
	return false
}

func (bi *baseItem) completeCut() item {
	return nil
}

func (bi *baseItem) canWash() bool {
	return false
}

func (bi *baseItem) completeWash() item {
	return nil
}

type cleanPlate struct {
	baseItem
	contents map[string]item
}

func newCleanPlate(items []item) *cleanPlate {
	cp := &cleanPlate{
		baseItem: baseItem{
			img: &cleanPlateImage,
		},
		contents: map[string]item{},
	}
	for _, i := range items {
		if !cp.combine(i) {
			panic("Not allowed to combine those items")
		}
	}
	return cp
}

func (p *cleanPlate) draw(screen *ebiten.Image, x, y float64) {
	p.baseItem.draw(screen, x, y)

	items := make([]item, 0, len(p.contents))
	for _, i := range p.contents {
		items = append(items, i)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].name() < items[j].name()
	})
	xOff := 1
	yOff := 0
	for _, i := range items {
		i.draw(screen, x+(float64(xOff)*8), y+(float64(yOff)*8))
		xOff++
		if xOff > 1 {
			xOff = 0
			yOff = 1
		}
	}
}

func (p *cleanPlate) combine(other item) bool {
	if !other.canCombine() {
		return false
	}

	if len(p.contents) >= 3 {
		return false
	}

	if _, ok := p.contents[other.name()]; ok {
		return false
	}

	p.contents[other.name()] = other
	return true
}

func (p *cleanPlate) canCombine() bool {
	return false
}

func (p *cleanPlate) name() string {
	return "cleanPlate"
}

type dirtyPlate struct {
	baseItem
}

func newDirtyPlate() *dirtyPlate {
	return &dirtyPlate{
		baseItem: baseItem{
			img: &dirtyPlateImage,
		},
	}
}

func (dp *dirtyPlate) canCombine() bool {
	return false
}

func (dp *dirtyPlate) name() string {
	return "dirtyPlate"
}

func (dp *dirtyPlate) canWash() bool {
	return true
}

func (dp *dirtyPlate) completeWash() item {
	return newCleanPlate(nil)
}

type lettuce struct {
	baseItem
}

func newLettuce() *lettuce {
	return &lettuce{
		baseItem: baseItem{
			img: &lettuceImage,
		},
	}
}

func (l *lettuce) canCombine() bool {
	return false
}

func (l *lettuce) name() string {
	return "lettuce"
}

func (l *lettuce) canCut() bool {
	return true
}

func (l *lettuce) completeCut() item {
	return newChoppedLettuce()
}

type choppedLettuce struct {
	baseItem
}

func newChoppedLettuce() *choppedLettuce {
	return &choppedLettuce{
		baseItem: baseItem{
			img: &choppedLettuceImage,
		},
	}
}

func (cl *choppedLettuce) canCombine() bool {
	return true
}

func (cl *choppedLettuce) name() string {
	return "choppedLettuce"
}

type tomato struct {
	baseItem
}

func newTomato() *tomato {
	return &tomato{
		baseItem: baseItem{
			img: &tomatoImage,
		},
	}
}

func (t *tomato) canCombine() bool {
	return false
}

func (t *tomato) name() string {
	return "tomato"
}

func (t *tomato) canCut() bool {
	return true
}

func (t *tomato) completeCut() item {
	return newChoppedTomato()
}

type choppedTomato struct {
	baseItem
}

func newChoppedTomato() *choppedTomato {
	return &choppedTomato{
		baseItem: baseItem{
			img: &choppedTomatoImage,
		},
	}
}

func (ct *choppedTomato) canCombine() bool {
	return true
}

func (ct *choppedTomato) name() string {
	return "choppedTomato"
}
