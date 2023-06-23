package main

import (
	"context"
	"errors"
	"fmt"
	"imagelnk2/internal/core"
	"imagelnk2/internal/site"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
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

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"version": "1.0.0",
		})
	})

	r.GET("/get", func(c *gin.Context) {
		url := c.Query("url")
		if url != "" {
			fmt.Printf("Process %s\n", url)

			retry := 0
			for retry < 5 {
				result, err := processor.GetImageURLs(url)
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

					log.Printf("%v", err)
					break
				} else {
					c.JSON(200, result)
					return
				}
			}
		}

		c.JSON(200, gin.H{})
	})

	r.Run(fmt.Sprintf("%s:%d", core.Config.Hostname, core.Config.Port))
}
