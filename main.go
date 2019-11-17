package main

import (
	"github.com/labstack/echo"
)

func main() {
	listen :="192.168.1.16:8080"
	e := echo.New()
	e.File("/", "public/index.html")
	e.Static("/download", "file")
	e.Static("/example", "example")
	e.Logger.Fatal(listen)
}

