package logic

import (
	"fmt"
	"math"
	"math/rand"
)

type GamePosition struct {
	X, Y int
}

type GameFood struct {
	Position GamePosition
}

type Mode int

const (
	ModeNotInitialized Mode = iota
	ModeNew
	ModeRunning
	ModeFinished
)

type Direction int

const (
	DirectionUp Direction = 1 + iota
	DirectionRight
	DirectionDown
	DirectionLeft
)

func (a GamePosition) getDistance(b GamePosition) float64 {
	dx := float64(b.X) - float64(a.X)
	dy := float64(b.Y) - float64(a.Y)
	return math.Sqrt(dx*dx + dy*dy)
}

func getRandomDirection() Direction {
	randomint := rand.Intn(4) + 1
	return Direction(randomint)
}

// Assumes that lines are horizontal or vertical
func checkLinesCrossed(linea1, linea2, lineb1, lineb2 GamePosition) bool {
	// Ensure linea1 is the top or left point, and linea2 is the bottom or right point
	if linea1.X > linea2.X || linea1.Y > linea2.Y {
		linea1, linea2 = linea2, linea1
	}

	// Ensure lineb1 is the top or left point, and lineb2 is the bottom or right point
	if lineb1.X > lineb2.X || lineb1.Y > lineb2.Y {
		lineb1, lineb2 = lineb2, lineb1
	}

	// Check if lines are vertical and horizontal
	isLineaVertical := linea1.X == linea2.X
	isLineaHorizontal := linea1.Y == linea2.Y
	isLinebVertical := lineb1.X == lineb2.X
	isLinebHorizontal := lineb1.Y == lineb2.Y

	if isLineaVertical && isLineaHorizontal {
		return false
	}

	if isLinebVertical && isLinebHorizontal {
		return false
	}

	// If one line is vertical and the other is horizontal, check for intersection
	if isLineaVertical && isLinebHorizontal {
		// Linea is vertical and Lineb is horizontal
		crossed := lineb1.X < linea1.X && lineb2.X > linea1.X &&
			linea1.Y < lineb1.Y && linea2.Y > lineb1.Y
		if crossed {
			fmt.Println("Crossed: ", linea1, linea2, lineb1, lineb2)
		}
		return crossed
	} else if isLineaHorizontal && isLinebVertical {
		// Linea is horizontal and Lineb is vertical
		crossed := linea1.X < lineb1.X && linea2.X > lineb1.X &&
			lineb1.Y < linea1.Y && lineb2.Y > linea1.Y
		if crossed {
			fmt.Println("Crossed: ", linea1, linea2, lineb1, lineb2)
		}
		return crossed
	}

	// If both lines are either vertical or horizontal, they can't cross
	return false
}
