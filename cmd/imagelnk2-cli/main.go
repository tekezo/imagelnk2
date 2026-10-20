package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"imagelnk2/internal/core"
	"imagelnk2/internal/processor"
	"log"
	"os"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

const maxAttempts = 5

type imageURLProcessor interface {
	GetImageURLs(context.Context, string) (*core.Result, error)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s URL\n", os.Args[0])
	}
	flag.Parse()

	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}

	if err := run(flag.Arg(0)); err != nil {
		log.Printf("error: %v", err)
		os.Exit(1)
	}
}

func run(url string) error {
	if err := core.ReadConfig(); err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	controlURL := launcher.New().Bin("/opt/google/chrome/chrome").MustLaunch()
	browser := rod.New().ControlURL(controlURL).MustConnect()
	defer browser.MustClose()

	result, err := getWithRetry(context.Background(), processor.New(browser), url)
	if err != nil {
		return err
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		return fmt.Errorf("encode result: %w", err)
	}

	return nil
}

func getWithRetry(ctx context.Context, p imageURLProcessor, url string) (*core.Result, error) {
	var lastErr error

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		result, err := p.GetImageURLs(ctx, url)
		if err == nil {
			return result, nil
		}

		lastErr = err
		var mandatoryElementErr *core.ErrMandatoryElementNotFound
		if !errors.Is(err, context.DeadlineExceeded) && !errors.As(err, &mandatoryElementErr) {
			return nil, err
		}

		if attempt < maxAttempts {
			log.Printf("retry %d/%d: %v", attempt, maxAttempts, err)
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", maxAttempts, lastErr)
}
