//go:build !js && !windows
// +build !js,!windows

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/nikpivkin/snake/game"
	"golang.org/x/sys/unix"
)

const (
	escape = 27
	space  = 32
)

const (
	gameTickDelay = 500 * time.Millisecond
)

var keyBindings = map[rune]game.Direction{
	'w': game.Up,
	'a': game.Left,
	's': game.Down,
	'd': game.Right,
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	errs := make(chan error)

	hideCursor()
	defer showCursor()
	clearTerminal()

	g := game.NewGame()
	g.Start()
	drawGame(g)

	key := make(chan rune, 1)
	go readKey(ctx, cancel, key, errs)
	go gameLoop(ctx, g, key, errs)

	select {
	case <-ctx.Done():
		fmt.Println("Exiting...")
		cancel()
		return nil
	case err := <-errs:
		return err
	}
}

func readKey(ctx context.Context, cancel context.CancelFunc, key chan<- rune, errs chan<- error) {
	fd := int(os.Stdin.Fd())
	oldState, err := makeRaw(fd)
	if err != nil {
		errs <- err
		return
	}

	// reset terminal on exit
	defer func() {
		_ = unix.IoctlSetTermios(fd, unix.TIOCSETA, oldState)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		r, ok, err := kbhit()
		if err != nil {
			errs <- err
			return
		}

		if !ok {
			continue
		}

		if r == escape {
			cancel()
			break
		}

		key <- r
	}
}

func gameLoop(ctx context.Context, g *game.Game, key <-chan rune, errs chan<- error) {
	gameTicker := time.NewTicker(gameTickDelay)
	defer gameTicker.Stop()

	var pause bool

	for {
		select {
		case <-ctx.Done():
			return
		case <-gameTicker.C:
			select {
			case k := <-key:

				if k == space {
					pause = !pause
				}

				if dir, ok := keyBindings[k]; ok {
					g.Move(dir)
				}
			default:
			}

			if !pause {
				g.Tick()

				if g.IsOver() {
					errs <- errors.New("game over")
					return
				} else if g.IsWin() {
					fmt.Println("YOU WIN")
					return
				}
			}

			clearTerminal()
			if pause {
				fmt.Println("Pause. Press space to continue.")
				continue
			}
			drawGame(g)
		}
	}
}

func kbhit() (rune, bool, error) {

	b := make([]byte, 4)
	n, err := os.Stdin.Read(b)

	if err != nil {
		return 0, false, err
	}

	r, sz := utf8.DecodeRune(b)

	if r == utf8.RuneError && sz == 1 {
		return 0, false, nil
	}

	return r, n == 1, nil
}

func hideCursor() {
	os.Stdout.WriteString("\033[?25l")
}

func showCursor() {
	os.Stdout.WriteString("\033[?25h")
}

func clearTerminal() {
	os.Stdout.WriteString("\033[H\033[2J")
}

func drawGame(g *game.Game) {
	var sb strings.Builder

	sb.WriteString("Score : ")
	sb.WriteString(strconv.Itoa(g.Score()))
	sb.WriteRune('\t')
	sb.WriteString("Length : ")
	sb.WriteString(strconv.Itoa(g.Length()))
	sb.WriteRune('\n')
	sb.WriteRune('\n')

	g.Walk(func(x, y int, c game.CellType) {
		switch c {
		case game.CellEmpty:
			sb.WriteRune('.')
		case game.CellFood:
			sb.WriteString("\033[33m")
			sb.WriteRune('F')
			sb.WriteString("\033[0m")
		case game.CellSnake:
			if g.IsHead(x, y) {
				switch g.Direction() {
				case game.Up:
					sb.WriteRune('^')
				case game.Right:
					sb.WriteRune('>')
				case game.Down:
					sb.WriteRune('v')
				case game.Left:
					sb.WriteRune('<')
				default:
					sb.WriteRune('x')
				}
			} else {
				sb.WriteRune('o')
			}
		}

		if x == game.BoardSize-1 {
			sb.WriteRune('\n')
		}
	})

	os.Stdout.WriteString(sb.String())
}

func makeRaw(fd int) (*unix.Termios, error) {
	var oldState *unix.Termios
	tio, err := unix.IoctlGetTermios(fd, unix.TIOCGETA)
	if err != nil {
		return nil, err
	}
	oldState = tio

	newState := oldState
	// Устанавливаем флаги для режима безбуферного ввода

	// https://man7.org/linux/man-pages/man3/termios.3.html
	// ICANON - enables canonical input processing mode
	// ECHO - enables echoing of input characters
	newState.Lflag &^= unix.ICANON | unix.ECHO
	if err := unix.IoctlSetTermios(fd, unix.TIOCSETA, newState); err != nil {
		return nil, err
	}
	return oldState, nil

}
