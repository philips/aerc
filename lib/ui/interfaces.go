package ui

import (
	"github.com/gdamore/tcell"
)

// Drawable is a UI component that can draw. Unless specified, all methods must
// only be called from a single goroutine, the UI goroutine.
type Drawable interface {
	// Called when this renderable should draw itself.
	Draw(ctx *Context)
	// Specifies a function to call when this cell needs to be redrawn. The
	// callback may be called in any goroutine.
	OnInvalidate(callback func(d Drawable))
	// Invalidates the drawable. This can be called from any goroutine.
	Invalidate()
}

type Interactive interface {
	// Returns true if the event was handled by this component
	Event(event tcell.Event) bool
	// Indicates whether or not this control will receive input events
	Focus(focus bool)
}

type Simulator interface {
	// Queues up the given input events for simulation
	Simulate(events []tcell.Event)
}

type DrawableInteractive interface {
	Drawable
	Interactive
}

// A drawable which contains other drawables
type Container interface {
	Drawable
	// Return all of the drawables which are children of this one (do not
	// recurse into your grandchildren).
	Children() []Drawable
}
