package main

import (
	"encoding/json"
	"fmt"
	"imagelnk2/internal/core"
	"imagelnk2/internal/debug"
	"imagelnk2/internal/processor"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

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

	controlURL := launcher.New().Bin("/opt/google/chrome/chrome").
		Set("lang", core.Config.Lang).
		MustLaunch()
	browser := rod.New().ControlURL(controlURL).MustConnect()
	defer browser.MustClose()

	processor := processor.New(browser)

	if len(os.Args) < 2 {
		fmt.Println("Usage: debug_run testdata/single.jsonc")
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
		log.Printf("check %s", debugEntry.Filename)

		result, err := processor.Debug(debugEntry.URL, debugEntry.Filename)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("%v", result)

		//
		// Test Title
		//

		{
			expected := debugEntry.Result.Title
			actual := result.Title

			if strings.HasPrefix(expected, "regexp:") {
				r := regexp.MustCompile(strings.TrimPrefix(expected, "regexp:"))
				if !r.MatchString(actual) {
					log.Fatalf("`%v` is not matched `%v`", actual, r)
				}
			} else {
				if expected != actual {
					log.Fatalf("`%v` != `%v`", expected, actual)
				}
			}
		}

		//
		// Test ImageURLs
		//

		if len(debugEntry.Result.ImageURLs) != len(result.ImageURLs) {
			log.Fatalf("`%v` != `%v`", debugEntry.Result.ImageURLs, result.ImageURLs)
		}

		for i := 0; i < len(debugEntry.Result.ImageURLs); i++ {
			expected := debugEntry.Result.ImageURLs[i]
			actual := result.ImageURLs[i]

			if strings.HasPrefix(expected, "regexp:") {
				r := regexp.MustCompile(strings.TrimPrefix(expected, "regexp:"))
				if !r.MatchString(actual) {
					log.Fatalf("`%v` is not matched `%v`", actual, r)
				}
			} else {
				if expected != actual {
					log.Fatalf("`%v` != `%v`", expected, actual)
				}
			}
		}

		//
		// Test ExtraURLs
		//

		if !reflect.DeepEqual(debugEntry.Result.ExtraURLs, result.ExtraURLs) {
			log.Fatalf("`%v` != `%v`", debugEntry.Result.ExtraURLs, result.ExtraURLs)
		}
	}
}
