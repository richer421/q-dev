package main

import (
	"q-dev/infra/mysql/model"

	"gorm.io/gen"
)

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath: "./infra/mysql/dao",
		Mode:    gen.WithDefaultQuery,
	})

	g.ApplyBasic(model.HelloWorld{})

	g.Execute()
}
