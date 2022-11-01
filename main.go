package main

import (
	"os"
	"scrapper/scrapper"
	"strings"

	"github.com/labstack/echo/v4"
)

func handleHome(c echo.Context) error {
	return c.File("home.html")
}

func handleScrape(c echo.Context) error {
	defer os.Remove("results.csv")
	term := strings.ToLower(scrapper.TrimSpace(c.FormValue("term")))
	scrapper.Scrape(term)
	return c.Attachment("results.csv", "results.csv")
}

func main() {
	e := echo.New()

	e.GET("/", handleHome)
	e.POST("/scrape", handleScrape)

	e.Logger.Fatal(e.Start(":1323"))
}
