package logic

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const FoodSize = 10

type UpdateFunc func(game *SnakeGame) error

type SnakeGame struct {
	mu             sync.RWMutex
	Height, Width  int
	Food           []GameFood
	Snake          []GamePosition // First is the head of the snake
	snakeLength    int
	snakeDirection Direction
	lastTurn       time.Time // Timestamp of setting snake[1]
	Mode           Mode
	onUpdates      map[int]UpdateFunc
	nextUpdateId   int
}

func (eng *SnakeGame) AddUpdateFunc(onUpdate UpdateFunc) int {
	eng.mu.Lock()
	defer eng.mu.Unlock()
	eng.nextUpdateId++
	if eng.onUpdates == nil {
		eng.onUpdates = map[int]UpdateFunc{}
	}
	eng.onUpdates[eng.nextUpdateId] = onUpdate
	return eng.nextUpdateId
}

func (eng *SnakeGame) RemoveUpdateFunc(id int) {
	eng.mu.Lock()
	defer eng.mu.Unlock()
	delete(eng.onUpdates, id)
}

func (eng *SnakeGame) SetSnakeDirection(direction Direction) {
	// Check, if change is required
	eng.mu.RLock()
	isNewDirection := eng.snakeDirection != direction
	eng.mu.RUnlock()
	if !isNewDirection {
		return
	}

	eng.mu.Lock()
	eng.snakeDirection = direction
	eng.Snake = append([]GamePosition{{X: eng.Snake[0].X, Y: eng.Snake[0].Y}}, eng.Snake...)
	eng.lastTurn = time.Now()
	eng.mu.Unlock()
}

func (eng *SnakeGame) GetRandomPosition() GamePosition {
	return GamePosition{
		X: rand.Intn(eng.Width),
		Y: rand.Intn(eng.Height),
	}
}

func (eng *SnakeGame) Restart(width, height, foodCount int) {
	eng.mu.Lock()
	eng.Mode = ModeNew
	eng.Width = width
	eng.Height = height
	snakePosition := eng.GetRandomPosition()
	eng.Snake = []GamePosition{snakePosition, snakePosition} // Start and end are the same
	eng.snakeDirection = getRandomDirection()
	eng.snakeLength = 40 // Start length of snake
	eng.Food = make([]GameFood, foodCount)
	for index := range eng.Food {
		eng.Food[index] = GameFood{
			Position: eng.GetRandomPosition(),
		}
	}
	eng.mu.Unlock()
}

// Only works because snake only has vertical and horisontal parts
func snakePartLength(a, b GamePosition) int {
	dx := b.X - a.X
	if dx < 0 {
		dx = -dx
	}
	dy := b.Y - a.Y
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func (eng *SnakeGame) getRealSnakeLength() int {
	realLength := 0
	for index, part := range eng.Snake {
		if index == 0 {
			continue
		}
		realLength += snakePartLength(part, eng.Snake[index-1])
	}
	return realLength
}

const FPS = 120
const sleepInterval = time.Second / FPS
const snakeSpeed float64 = 20.0 / (1.0 * float64(time.Second))

// Run this in a separate thread using "go engine.Run()"
func (eng *SnakeGame) Run(ctx context.Context) error {

	for {
		// Wait for the game to start
		for eng.Mode != ModeNew {
			time.Sleep(sleepInterval)
		}

		// Start the game
		eng.mu.Lock()
		eng.lastTurn = time.Now()
		eng.Mode = ModeRunning
		eng.mu.Unlock()

		// Run the game
		for eng.Mode == ModeRunning {
			time.Sleep(sleepInterval)
			eng.mu.Lock()

			// Move snake head
			durSinceTurn := time.Since(eng.lastTurn)
			distanceSinceTurnFloat := float64(durSinceTurn) * snakeSpeed
			distanceSinceTurn := int(distanceSinceTurnFloat)
			switch eng.snakeDirection {
			case DirectionDown:
				eng.Snake[0].Y = eng.Snake[1].Y + distanceSinceTurn
			case DirectionUp:
				eng.Snake[0].Y = eng.Snake[1].Y - distanceSinceTurn
			case DirectionRight:
				eng.Snake[0].X = eng.Snake[1].X + distanceSinceTurn
			case DirectionLeft:
				eng.Snake[0].X = eng.Snake[1].X - distanceSinceTurn
			default:
				panic("Illegal direction")
			}

			// Detect if snake hits itself
			for segindex := range eng.Snake {
				if segindex <= 1 {
					continue
				}
				if checkLinesCrossed(eng.Snake[0], eng.Snake[1], eng.Snake[segindex-1], eng.Snake[segindex]) {
					fmt.Println("Snake crossed itself")
					eng.Mode = ModeFinished
				}

			}

			// Detect if food was hit
			for foodindex, food := range eng.Food {
				if food.Position.getDistance(eng.Snake[0]) < FoodSize {
					fmt.Println("Food found!!")
					eng.Food[foodindex].Position = eng.GetRandomPosition()
					eng.snakeLength = eng.snakeLength * 3 / 2
				}
			}

			// Insert extra point in case the snake has no turns, so that lastTurn time has a well defined location
			if len(eng.Snake) == 2 {
				eng.Snake = []GamePosition{eng.Snake[0], eng.Snake[0], eng.Snake[1]}
				eng.lastTurn = time.Now()
			}

			// Move snake end
			realLength := eng.getRealSnakeLength()
			extraLength := realLength - eng.snakeLength
			if extraLength > 0 {
				snakeLastIndex := len(eng.Snake) - 1
				snakeLastPartLength := snakePartLength(eng.Snake[snakeLastIndex], eng.Snake[snakeLastIndex-1])
				if snakeLastPartLength <= extraLength {
					eng.Snake = eng.Snake[0:snakeLastIndex]
				} else {
					if eng.Snake[snakeLastIndex].X > eng.Snake[snakeLastIndex-1].X {
						eng.Snake[snakeLastIndex].X -= extraLength
					} else if eng.Snake[snakeLastIndex].X < eng.Snake[snakeLastIndex-1].X {
						eng.Snake[snakeLastIndex].X += extraLength
					} else if eng.Snake[snakeLastIndex].Y > eng.Snake[snakeLastIndex-1].Y {
						eng.Snake[snakeLastIndex].Y -= extraLength
					} else if eng.Snake[snakeLastIndex].Y < eng.Snake[snakeLastIndex-1].Y {
						eng.Snake[snakeLastIndex].Y += extraLength
					} else {
						// Snake's back part has length 0 so we remove it. Next iteration will then cut parts of the next section
						eng.Snake = eng.Snake[0:snakeLastIndex]
					}
				}
			}

			// Check if snake has hit the wall
			if eng.Snake[0].X < 0 || eng.Snake[0].X > eng.Width || eng.Snake[0].Y < 0 || eng.Snake[0].Y > eng.Height {
				fmt.Println("Snake hit the wall")
				eng.Mode = ModeFinished
			}

			eng.mu.Unlock()

			onUpdateCount := len(eng.onUpdates)
			if onUpdateCount > 0 {
				wg := sync.WaitGroup{}
				wg.Add(onUpdateCount)

				errs := make([]error, 0, onUpdateCount)
				for _, onUpdate := range eng.onUpdates {
					go func(onUpdate UpdateFunc) {
						defer wg.Done()
						if err := onUpdate(eng); err != nil {
							errs = append(errs, err)
						}
					}(onUpdate)
				}
				wg.Wait()
			}
		}
	}
}
