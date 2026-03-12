package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/KooEric/prtr/internal/app"
	"github.com/KooEric/prtr/internal/clipboard"
	"github.com/KooEric/prtr/internal/config"
	"github.com/KooEric/prtr/internal/input"
	"github.com/KooEric/prtr/internal/translate"
)

var version = "dev"

func main() {
	stdinPiped, err := input.StdinIsPiped(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to inspect stdin: %v\n", err)
		os.Exit(1)
	}

	application := app.New(app.Dependencies{
		Version: version,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
		LookupEnv: func(key string) (string, bool) {
			return os.LookupEnv(key)
		},
		ConfigLoader: config.Load,
		ConfigInit:   config.Init,
		Translator: translate.NewDeepLClient(translate.ClientOptions{
			APIKey:  os.Getenv("DEEPL_API_KEY"),
			BaseURL: translate.DefaultBaseURL,
			HTTPClient: &http.Client{
				Timeout: 15 * time.Second,
			},
		}),
		Clipboard: clipboard.NewPBClipboard(),
	})

	if err := application.Execute(context.Background(), os.Args[1:], os.Stdin, stdinPiped); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
