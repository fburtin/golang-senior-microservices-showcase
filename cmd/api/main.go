package main

import "github.com/fburtin/golang-senior-microservices-showcase/internal/app"

func main() {
	application := app.New()
	application.Run()
}
