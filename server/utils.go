package server

import (
	"math/rand"
)

func newState(width, height int) state {
	return state{
		GameID: "",
		Width:  width,
		Height: height,
		Fruit:  newFruit(width, height, 0, 0),
		Snake:  newSnake(),
	}
}

func newSnake() snake {
	return updateSnake(0, 0, 1, 0)
}

func updateSnake(x, y, velX, velY int) snake {
	return snake{
		X:    x,
		Y:    y,
		VelX: velX,
		VelY: velY,
	}
}

// for new state, (curX, curY) is (0, 0), since it is snake's initial position
func newFruit(width, height, curX, curY int) fruit {
	generated := fruit{
		X: rand.Intn(width),
		Y: rand.Intn(height),
	}

	if generated.X == curX && generated.Y == curY {
		generated = newFruit(width, height, curX, curY)
	}

	return generated
}
