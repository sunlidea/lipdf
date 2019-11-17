package main

import (
	"github.com/labstack/echo"
	"github.com/sunlidea/lipdf/handler"
)

func main() {
	e := echo.New()

	//static file
	e.Static("/", "public")
	e.Static("/download", "file")
	e.Static("/example", "example")

	wh := &handler.WebHandler{}
	//handler
	e.POST("/submit", wh.Submit)
	e.POST("/example", wh.Example)
	e.POST("/upload", wh.Upload)

	e.Logger.Fatal(e.Start(":1323"))
}

