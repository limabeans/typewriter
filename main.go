package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell"
)

var (
	screen   tcell.Screen
	events   chan tcell.Event
	defStyle tcell.Style
	cur      int
	buffer   []rune
	f        *os.File
)

func ResetScreen() {
	w, h := screen.Size()
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			screen.SetContent(x, y, ' ', nil, defStyle)
		}
	}
	screen.Show()
}

func RefreshScreen() {

	f.WriteString("RefreshScreen\n")
	f.Sync()

	w, h := screen.Size()
	col := cur % w
	row := cur / w
	screen.ShowCursor(col, row)
	for i := 0; i < w*h; i++ {
		col := i % w
		row := i / w
		screen.SetContent(col, row, buffer[i], nil, defStyle)
	}
	screen.Show()
	col = cur % w
	row = cur / w
	screen.ShowCursor(col, row)
}

func insertChar(c rune) {
	w, h := screen.Size()
	// shift everything from cur to end down
	for i := w*h - 1; i >= cur; i-- {
		if i == 0 {
			continue
		}
		buffer[i] = buffer[i-1]
	}

	buffer[cur] = c

	cur++
}

func arrowLeft() {
	if cur == 0 {
		return
	}
	cur--
}
func arrowRight() {
	w, h := screen.Size()
	if cur == (w*h)-1 {
		return
	}
	cur++
}

func arrowDown() {
	w, h := screen.Size()
	tmp := cur + w
	if tmp > (w*h)-1 {
		return
	}
	cur = tmp
}

func arrowUp() {
	w, _ := screen.Size()
	tmp := cur - w
	if tmp < 0 {
		return
	}
	cur = tmp
}

func main() {
	var err error
	// init tcell screen
	screen, err = tcell.NewScreen()
	defer screen.Fini()
	if err != nil {
		os.Exit(1)
	}
	if err = screen.Init(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cur = 0

	// init buffer slice
	w, h := screen.Size()
	buffer = make([]rune, w*h)
	for i := range buffer {
		buffer[i] = ' '
	}

	// events channel
	events = make(chan tcell.Event, 100)

	// go routine for polling kbd events
	go func() {
		for {
			if screen != nil {
				events <- screen.PollEvent()
			}
		}
	}()

	// For debugging, use `tail -f kbdlog`
	f, _ = os.Create("kbdlog")
	defer f.Close()

	for {
		RefreshScreen()

		var event tcell.Event
		event = <-events
		if event != nil {
			switch e := event.(type) {

			case *tcell.EventKey:
				if e.Modifiers() == 2 && e.Rune() == 17 {
					// TODO: Look into defer, because I still need to call Fini() here before exiting.
					screen.Fini()
					os.Exit(0)
				} else if e.Key() == tcell.KeyLeft || (e.Modifiers() == 2 && e.Rune() == 2) {
					arrowLeft()
				} else if e.Key() == tcell.KeyRight || (e.Modifiers() == 2 && e.Rune() == 6) {
					arrowRight()
				} else if e.Key() == tcell.KeyUp || (e.Modifiers() == 2 && e.Rune() == 16) {
					arrowUp()
				} else if e.Key() == tcell.KeyDown || (e.Modifiers() == 2 && e.Rune() == 14) {
					arrowDown()
				} else {
					insertChar(e.Rune())
				}

				f.WriteString(formatEventKey(e))
				f.Sync()
				event = nil
			}
		}
	}
}

func formatEventKey(e *tcell.EventKey) string {
	mod := e.Modifiers()
	key := e.Key()
	rune := e.Rune()

	return fmt.Sprintf("{mod: %v, key: %v, rune: %v}\n", mod, key, rune)
}
