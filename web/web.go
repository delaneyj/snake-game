package web

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/delaneyj/snake/logic"
	"github.com/go-chi/chi"
	datastar "github.com/starfederation/datastar/sdk/go"
)

const port = 8080

func RunHTTPServer(ctx context.Context, sharedGame *logic.SnakeGame) error {

	r := chi.NewRouter()

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}

	foodSize := 10

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		PageSnake(sharedGame, foodSize).Render(ctx, w)
	})

	r.Get("/updates", func(w http.ResponseWriter, r *http.Request) {
		sse := datastar.NewSSE(w, r)
		updateID := sharedGame.AddUpdateFunc(func(game *logic.SnakeGame) error {
			if err := errors.Join(
				sse.MergeFragmentTempl(SnakeArenaSVG(game, foodSize)),
				sse.MergeFragmentTempl(SnakeButtons(sharedGame)),
			); err != nil {
				return fmt.Errorf("rendering fragments: %w", err)
			}

			return nil
		})
		defer sharedGame.RemoveUpdateFunc(updateID)
		<-r.Context().Done()
	})

	r.Post("/reset", func(w http.ResponseWriter, r *http.Request) {
		sharedGame.Restart(600, 300, 10)
		datastar.NewSSE(w, r)
	})

	r.Post("/inputs/{direction}", func(w http.ResponseWriter, r *http.Request) {
		direction := chi.URLParam(r, "direction")
		switch direction {
		case "up":
			sharedGame.SetSnakeDirection(logic.DirectionUp)
		case "down":
			sharedGame.SetSnakeDirection(logic.DirectionDown)
		case "left":
			sharedGame.SetSnakeDirection(logic.DirectionLeft)
		case "right":
			sharedGame.SetSnakeDirection(logic.DirectionRight)
		default:
			http.Error(w, "Invalid direction", http.StatusBadRequest)
			return
		}

		datastar.NewSSE(w, r)
	})

	go func() {
		<-ctx.Done()
		srv.Shutdown(ctx)
	}()

	log.Printf("Listening on http://localhost:%d", port)
	return srv.ListenAndServe()
}

func position2points(positions []logic.GamePosition) string {
	sb := strings.Builder{}
	for _, p := range positions {
		sb.WriteString(fmt.Sprintf("%d,%d ", p.X, p.Y))
	}
	snakeParts := sb.String()
	return snakeParts
}
