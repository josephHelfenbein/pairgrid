package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func handle() *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Welcome to Vercel!")
	})

	api := e.Group("/api")
	api.GET("/users", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Users endpoint"})
	})

	api.GET("/test", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{"message": "Test endpoint"})
	})
	return e
}
func Handler(w http.ResponseWriter, r *http.Request) {
	e := handle()

	e.ServeHTTP(w, r)
}

func main() {
	e := handle()

	e.Logger.Fatal(e.Start(":8080"))
}
