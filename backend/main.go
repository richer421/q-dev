package main

import (
	"{{ .ModuleName }}/cmd"
)

//go:generate go run github.com/swaggo/swag/cmd/swag init -o ./gen/docs --parseDependency
//go:generate go run ./gen/gorm_gen

func main() {
	cmd.Execute()
}
