package game

import "math/rand"

type (
	Direction uint8
	CellType  uint8
)

const (
	BoardSize = 10
)

const (
	Undefine Direction = iota
	Up
	Right
	Down
	Left
)

const (
	CellEmpty CellType = iota
	CellSnake
	CellFood
)

func (d Direction) isOppositeTo(other Direction) bool {
	if d == Undefine || other == Undefine {
		return false
	}
	return abs(int(d)-int(other)) == 2
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// snake Game
type Game struct {
	cells [BoardSize][BoardSize]CellType

	// snake
	snakeHeadX int
	snakeHeadY int
	snakeBody  [][2]int

	food      [2]int
	score     int
	direction Direction

	gameOver bool
	win      bool
}

func NewGame() *Game {
	return &Game{
		cells:     [BoardSize][BoardSize]CellType{},
		direction: Undefine,
	}
}

func (g *Game) Start() {

	g.snakeHeadX = rand.Intn(BoardSize)
	g.snakeHeadY = rand.Intn(BoardSize)

	g.growSnake()
	g.spawnFood()

}

func (g *Game) Move(dir Direction) {
	if !g.direction.isOppositeTo(dir) {
		g.direction = dir
	}
}

func (g *Game) IsOver() bool {
	return g.gameOver
}

func (g *Game) IsWin() bool {
	return g.win
}

func (g *Game) Score() int {
	return g.score
}

func (g *Game) Direction() Direction {
	return g.direction
}

func (g *Game) IsHead(x, y int) bool {
	return g.snakeHeadX == x && g.snakeHeadY == y
}

func (g *Game) Length() int {
	return len(g.snakeBody)
}

func (g *Game) Walk(fn func(x, y int, c CellType)) {
	for y := 0; y < BoardSize; y++ {
		for x := 0; x < BoardSize; x++ {
			fn(x, y, g.cells[x][y])
		}
	}
}

func (g *Game) Tick() {

	if g.direction == 0 {
		return
	}

	g.moveSnake()

	if g.wallCollision() || g.snakeCollision() {
		g.gameOver = true
		return
	}

	g.growSnake()

	if len(g.freeCells()) == 0 {
		g.win = true
		return
	}

	if g.foodCollision() {
		g.eatFood()
	} else {
		g.removeSnakeTail()
	}
}

func (g *Game) moveSnake() {
	switch g.direction {
	case Up:
		g.snakeHeadY--
	case Right:
		g.snakeHeadX++
	case Down:
		g.snakeHeadY++
	case Left:
		g.snakeHeadX--
	}
}

func (g *Game) wallCollision() bool {
	return g.snakeHeadX < 0 || g.snakeHeadX >= BoardSize || g.snakeHeadY < 0 || g.snakeHeadY >= BoardSize
}

func (g *Game) snakeCollision() bool {
	return g.cells[g.snakeHeadX][g.snakeHeadY] == CellSnake
}

func (g *Game) foodCollision() bool {
	return g.snakeHeadX == g.food[0] && g.snakeHeadY == g.food[1]
}

func (g *Game) removeSnakeTail() {
	g.cells[g.snakeBody[0][0]][g.snakeBody[0][1]] = CellEmpty
	g.snakeBody = g.snakeBody[1:]
}

func (g *Game) growSnake() {
	g.snakeBody = append(g.snakeBody, [2]int{g.snakeHeadX, g.snakeHeadY})
	g.cells[g.snakeHeadX][g.snakeHeadY] = CellSnake
}

func (g *Game) eatFood() {
	g.score += 5
	g.spawnFood()
}

func (g *Game) freeCells() [][2]int {
	var freeCells [][2]int
	for i := 0; i < BoardSize; i++ {
		for j := 0; j < BoardSize; j++ {
			if g.cells[i][j] == CellEmpty {
				freeCells = append(freeCells, [2]int{i, j})
			}
		}
	}
	return freeCells
}

func (g *Game) spawnFood() {
	freeCells := g.randomFreeCell()
	if len(freeCells) == 0 {
		return
	}
	g.food = freeCells
	g.cells[g.food[0]][g.food[1]] = CellFood
}

func (g *Game) randomFreeCell() [2]int {
	freeCells := g.freeCells()
	if len(freeCells) == 0 {
		return [2]int{}
	}
	return freeCells[rand.Intn(len(freeCells))]
}
