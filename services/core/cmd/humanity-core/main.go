// Command humanity-core serves the public read-only operational API.
// Local development binds to loopback by default; set ADDR only behind a deployment proxy.
package main

import (
	"context"
	"os"

	"github.com/Aravindh97/aravindh-impact-portfolio/internal/kit"
	"github.com/Aravindh97/aravindh-impact-portfolio/services/core/app"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = "127.0.0.1:8090"
	}
	if err := kit.Serve(addr, app.New(context.Background())); err != nil {
		panic(err)
	}
}
