package main

import "mxmz.it/mxmz/tommaso/dto"
import "github.com/labstack/echo/v4"

func main() {
	e := echo.New()
	api := e.Group("/api")
	api.POST("/my-probes", myProbes)
	e.Logger.Fatal(e.Start(":7997"))
}

func myProbes(c echo.Context) error {
	var err error
	var myNames dto.MyNames
	if err = c.Bind(&myNames); err != nil {
		return err
	}
	c.JSON(200, []string{"boh"})
	return nil
}
