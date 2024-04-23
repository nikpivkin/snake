//go:build js && wasm
// +build js,wasm

package main

import (
	"strconv"
	"syscall/js"

	"github.com/nikpivkin/snake/game"
)

const loopDelay = 500

// TODO: use tinygo
// https://github.com/tinygo-org/tinygo

var keyBindings = map[string]game.Direction{
	"w": game.Up,
	"a": game.Left,
	"s": game.Down,
	"d": game.Right,
}

func main() {

	var cleanup func()
	defer cleanup()

	js.Global().Call("whenLoaded", js.FuncOf(func(this js.Value, args []js.Value) any {
		cleanup = initGame()
		return nil
	}))

	<-make(chan struct{})
}

func initGame() func() {
	g := game.NewGame()
	g.Start()

	var key string

	js.Global().Get("document").Call("addEventListener", "keydown", js.FuncOf(func(this js.Value, args []js.Value) any {
		key = args[0].Get("key").String()
		js.Global().Get("console").Call("log", key)
		return nil
	}))

	d := newDrawer(g, 600, 400)
	d.draw()

	var loop js.Value
	loop = js.Global().Get("window").Call("setInterval", js.FuncOf(func(this js.Value, args []js.Value) any {

		if key == "Escape" {
			js.Global().Get("window").Call("clearInterval", loop)
			js.Global().Get("console").Call("log", "Exiting...")
			return nil
		}

		if k, ok := keyBindings[key]; ok {
			g.Move(k)
		}
		g.Tick()
		d.draw()

		if g.IsOver() || g.IsWin() {
			js.Global().Get("window").Call("clearInterval", loop)
			return nil
		}

		return nil
	}), loopDelay)

	return func() {
		js.Global().Call("console", "log", "Cleanup")
		js.Global().Get("window").Call("clearInterval", loop)
		d.stop()
	}
}

type drawer struct {
	game          *game.Game
	height, width int

	window js.Value
	ctx    js.Value
	reqID  js.Value
}

func newDrawer(game *game.Game, height, width int) *drawer {
	window := js.Global()
	doc := window.Get("document")
	canvas := doc.Call("createElement", "canvas")
	canvas.Set("width", width)
	canvas.Set("height", height)
	canvas.Set("id", "game")

	doc.Get("body").Call("appendChild", canvas)

	return &drawer{
		game:   game,
		height: height,
		width:  width,
		window: window,
		ctx:    canvas.Call("getContext", "2d"),
	}
}

func (d *drawer) draw() {
	d.ctx.Call("clearRect", 0, 0, d.width, d.height)

	d.ctx.Set("textAlign", "left")
	d.ctx.Set("fillStyle", "#fff")
	d.ctx.Set("font", "bold 20px Helvetica, Arial")
	d.ctx.Call("fillText", "Score : "+strconv.Itoa(d.game.Score()), 10, 20)
	d.ctx.Call("fillText", "Length : "+strconv.Itoa(d.game.Length()), 10, 40)

	cellSize := d.width / game.BoardSize

	// draw snake
	d.game.Walk(func(x, y int, c game.CellType) {
		d.ctx.Set("fillStyle", "#000")
		switch c {
		case game.CellEmpty:
		case game.CellFood:
			d.ctx.Set("fillStyle", "#f00")
		case game.CellSnake:
			if d.game.IsHead(x, y) {
				// dark green
				d.ctx.Set("fillStyle", "#173518")
			} else {
				d.ctx.Set("fillStyle", "#0f0")
			}
		}
		d.ctx.Call("fillRect", x*cellSize, (y*cellSize)+d.height-d.width, cellSize, cellSize)
	})

	d.ctx.Set("textAlign", "center")

	if d.game.IsOver() {
		d.ctx.Set("fillStyle", "#f00")
		d.ctx.Set("font", "bold 60px Helvetica, Arial")
		d.ctx.Call("fillText", "GAME OVER", d.width/2, d.height/2)
	} else if d.game.IsWin() {
		d.ctx.Set("fillStyle", "#f00")
		d.ctx.Set("font", "bold 60px Helvetica, Arial")
		d.ctx.Call("fillText", "YOU WIN", d.width/2, d.height/2)
	}

	if int(d.game.Direction()) == 0 {
		d.ctx.Set("fillStyle", "#fff")
		d.ctx.Set("font", "bold 30px Helvetica, Arial")
		d.ctx.Call("fillText", `Press w, a, s or d to start`, d.width/2, d.height/2)
	}
}

func (d *drawer) stop() {
	d.window.Call("cancelAnimationFrame", d.reqID)
}
