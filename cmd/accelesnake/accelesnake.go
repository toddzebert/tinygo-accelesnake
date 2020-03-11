package main

import (
	"image/color"
	"math/rand"
	"time"

	"../../assets/fonts"
	"tinygo.org/x/drivers/shifter"
	"tinygo.org/x/tinyfont"
)

const (
	BCK = iota
	SNAKE
	APPLE
	TEXT
)

const (
	START = iota
	BEGINPLAY
	PLAY
	GAMEOVER
	QUIT
)

const THRESHOLD int16 = 6

const SPEED = 150 // sleep time (ms) during gameplay

const (
	WIDTHBLOCKS  = 16
	HEIGHTBLOCKS = 13
)

type Snake struct {
	body      [208][2]int16
	length    int16
	direction int16
}

type Game struct {
	colors         []color.RGBA
	snake          Snake
	appleX, appleY int16
	status         uint8
}

func (game *Game) Start() {
	game.status = START
	display.FillScreen(game.colors[BCK])

	for {
		switch game.status {
			case START:
				game.stateStart()
				break

			case GAMEOVER:
				game.stateGameover()
				break

			case BEGINPLAY:
				game.stateBeginPlay()
				break

			case PLAY:
				game.statePlay()
				break
		}
	}
}


func (game *Game) stateStart() {
	display.FillScreen(game.colors[BCK])

	tinyfont.WriteLine(&display, &fonts.Bold24pt7b, 0, 50, []byte("SNAKE"), game.colors[TEXT])
	tinyfont.WriteLine(&display, &fonts.Regular12pt7b, 8, 100, []byte("Press START"), game.colors[TEXT])

	time.Sleep(2 * time.Second)

	for game.status == START {
		buttons.Read8Input()

		if buttons.Pins[BUTTON_START].Get() {
			game.status = BEGINPLAY
		}
	}
}


func (game *Game) stateGameover() {
	scoreStr := []byte("SCORE: 123")
	display.FillScreen(game.colors[BCK])

	scoreStr[7] = 48 + uint8((game.snake.length-3)/100)
	scoreStr[8] = 48 + uint8(((game.snake.length-3)/10)%10)
	scoreStr[9] = 48 + uint8((game.snake.length-3)%10)

	tinyfont.WriteLine(&display, &fonts.Regular12pt7b, 8, 50, []byte("GAME OVER"), game.colors[TEXT])
	tinyfont.WriteLine(&display, &fonts.Regular12pt7b, 8, 100, []byte("Press START"), game.colors[TEXT])
	tinyfont.WriteLine(&display, &tinyfont.TomThumb, 50, 120, scoreStr, game.colors[TEXT])

	time.Sleep(2 * time.Second)

	for game.status == GAMEOVER {
		buttons.Read8Input()

		if buttons.Pins[BUTTON_START].Get() {
			game.status = BEGINPLAY
		}
	}
}


func (game *Game) stateBeginPlay() {
	display.FillScreen(game.colors[BCK])
	game.snake.body = [208][2]int16{
		{0, 3},
		{0, 2},
		{0, 1},
		{0, 0},
	}
	game.snake.length = 4
	game.snake.direction = 3
	game.drawSnake()
	game.createApple()
	time.Sleep(2000 * time.Millisecond)
	game.status = PLAY
}

func (game *Game) statePlay() {
	x, y, z := accel.ReadRawAcceleration()
	x = x / 500
	y = y / 500
	z = z / 500

	if x < -THRESHOLD { // left
		if game.snake.direction != 3 {
			game.snake.direction = 0
		}
	} else if x > THRESHOLD { // right
		if game.snake.direction != 0 {
			game.snake.direction = 3
		}
	}

	if y < -THRESHOLD { // up
		if game.snake.direction != 2 {
			game.snake.direction = 1
		}
	} else if y > THRESHOLD { // down
		if game.snake.direction != 1 {
			game.snake.direction = 2
		}
	}

	if buttons.Pins[shifter.BUTTON_SELECT].Get() {
		game.status = START
	}

	game.moveSnake()
	time.Sleep(SPEED * time.Millisecond)
}


func (g *Game) collisionWithSnake(x, y int16) bool {
	for i := int16(0); i < g.snake.length; i++ {
		if x == g.snake.body[i][0] && y == g.snake.body[i][1] {
			return true
		}
	}
	return false
}


func (g *Game) createApple() {
	g.appleX = int16(rand.Int31n(16))
	g.appleY = int16(rand.Int31n(13))

	for g.collisionWithSnake(g.appleX, g.appleY) {
		g.appleX = int16(rand.Int31n(16))
		g.appleY = int16(rand.Int31n(13))
	}

	g.drawSnakePartial(g.appleX, g.appleY, g.colors[APPLE])
}


func (g *Game) moveSnake() {
	x := g.snake.body[0][0]
	y := g.snake.body[0][1]

	switch g.snake.direction {
		case 0:
			x--
			break
		case 1:
			y--
			break
		case 2:
			y++
			break
		case 3:
			x++
			break
	}

	if x >= WIDTHBLOCKS {
		x = 0
	} else if x < 0 {
		x = WIDTHBLOCKS - 1
	}

	if y >= HEIGHTBLOCKS {
		y = 0
	} else if y < 0 {
		y = HEIGHTBLOCKS - 1
	}

	if g.collisionWithSnake(x, y) {
		g.status = GAMEOVER
	}

	// draw head
	g.drawSnakePartial(x, y, g.colors[SNAKE])

	if x == g.appleX && y == g.appleY {
		g.snake.length++
		g.createApple()
	} else {
		// remove tail
		g.drawSnakePartial(g.snake.body[g.snake.length-1][0], g.snake.body[g.snake.length-1][1], g.colors[BCK])
	}

	for i := g.snake.length - 1; i > 0; i-- {
		g.snake.body[i][0] = g.snake.body[i-1][0]
		g.snake.body[i][1] = g.snake.body[i-1][1]
	}

	g.snake.body[0][0] = x
	g.snake.body[0][1] = y
}


func (g *Game) drawSnake() {
	for i := int16(0); i < g.snake.length; i++ {
		g.drawSnakePartial(g.snake.body[i][0], g.snake.body[i][1], g.colors[SNAKE])
	}
}


func (g *Game) drawSnakePartial(x, y int16, c color.RGBA) {
	modY := int16(9)

	if y == 12 {
		modY = 8
	}

	display.FillRectangle(10*x, 10*y, 9, modY, c)
}
