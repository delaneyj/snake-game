package main

import (
	"context"
	"log"

	"github.com/delaneyj/snake/logic"
	"github.com/delaneyj/snake/web"
	"github.com/delaneyj/toolbelt"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run(ctx context.Context) error {
	sharedGame := &logic.SnakeGame{}
	sharedGame.Restart(600, 300, 10)

	eg := toolbelt.NewErrGroupSharedCtx(
		ctx,
		func(ctx context.Context) error {
			return web.RunHTTPServer(ctx, sharedGame)
		},
		sharedGame.Run,
	)

	return eg.Wait()

}
