package ui

import (
	tb "github.com/nsf/termbox-go"
)

const (
	BORDER_LEFT   = 1 << iota
	BORDER_TOP    = 1 << iota
	BORDER_RIGHT  = 1 << iota
	BORDER_BOTTOM = 1 << iota
)

type Bordered struct {
	borders      uint
	content      Drawable
	onInvalidate func(d Drawable)
}

func NewBordered(content Drawable, borders uint) *Bordered {
	b := &Bordered{
		borders: borders,
		content: content,
	}
	content.OnInvalidate(b.contentInvalidated)
	return b
}

func (bordered *Bordered) contentInvalidated(d Drawable) {
	bordered.Invalidate()
}

func (bordered *Bordered) Invalidate() {
	if bordered.onInvalidate != nil {
		bordered.onInvalidate(bordered)
	}
}

func (bordered *Bordered) OnInvalidate(onInvalidate func(d Drawable)) {
	bordered.onInvalidate = onInvalidate
}

func (bordered *Bordered) Draw(ctx *Context) {
	x := 0
	y := 0
	width := ctx.Width()
	height := ctx.Height()
	cell := tb.Cell{
		Ch: ' ',
		Fg: tb.ColorBlack,
		Bg: tb.ColorWhite,
	}
	if bordered.borders&BORDER_LEFT != 0 {
		ctx.Fill(0, 0, 1, ctx.Height(), cell)
		x += 1
		width -= 1
	}
	if bordered.borders&BORDER_TOP != 0 {
		ctx.Fill(0, 0, ctx.Width(), 1, cell)
		y += 1
		height -= 1
	}
	if bordered.borders&BORDER_RIGHT != 0 {
		ctx.Fill(ctx.Width()-1, 0, 1, ctx.Height(), cell)
		width -= 1
	}
	if bordered.borders&BORDER_BOTTOM != 0 {
		ctx.Fill(0, ctx.Height()-1, ctx.Width(), 1, cell)
		height -= 1
	}
	subctx := ctx.Subcontext(x, y, width, height)
	bordered.content.Draw(subctx)
}