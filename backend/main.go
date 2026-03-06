package main

import (
	"q-dev/cmd"
)

//go:generate go run github.com/swaggo/swag/cmd/swag init -o ./gen/docs --parseDependency
//go:generate go run ./gen/gorm_gen

// @title       Q-Dev API
// @version     1.0
// @BasePath    /api
func main() {
	cmd.Execute()
}
