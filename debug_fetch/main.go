package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"imagelnk2/internal/core"
	"imagelnk2/internal/debug"
	"imagelnk2/internal/site"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"muzzammil.xyz/jsonc"
)

func main() {
	err := core.ReadConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	controlURL := launcher.New().Bin("/opt/google/chrome/chrome").MustLaunch()
	browser := rod.New().ControlURL(controlURL).MustConnect()
	defer browser.MustClose()

	processor := site.NewProcessor(browser)

	if len(os.Args) < 2 {
		fmt.Println("Usage: debug_fetch json")
		os.Exit(1)
	}

	//
	// gin
	//

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/testdata", func(c *gin.Context) {
		filename := c.Query("filename")

		if filename != "" {
			path := filepath.Join("testdata", "html", filepath.Base(filename))
			c.File(path)
			return
		}

		c.JSON(200, gin.H{})
	})

	go r.Run(fmt.Sprintf("%s:%d", core.Config.Hostname, core.Config.Port))

	//
	// load json
	//

	jsoncPath := os.Args[1]
	jsoncData, err := os.ReadFile(jsoncPath)
	if err != nil {
		log.Fatal(err)
	}

	jsonData := jsonc.ToJSON(jsoncData)

	var debugEntries []debug.Entry
	err = json.Unmarshal(jsonData, &debugEntries)
	if err != nil {
		log.Fatal(err)
	}

	for _, debugEntry := range debugEntries {
		retry := 0
		for retry < 5 {
			err := processor.SaveHTML(debugEntry.URL, debugEntry.Filename)
			if err != nil {
				var errMandatoryElementNotFound *core.ErrMandatoryElementNotFound
				if errors.Is(err, context.DeadlineExceeded) {
					log.Printf("timeout (retry %d)", retry)
					retry++
					continue
				}
				if errors.As(err, &errMandatoryElementNotFound) {
					log.Printf("%s (retry %d)", errMandatoryElementNotFound.Error(), retry)
					retry++
					continue
				}

				log.Println(err)
				break
			}

			break
		}
	}
}
