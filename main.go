package main

import (
	"github.com/aylei/alert-shim/pkg/api"
	"github.com/aylei/alert-shim/pkg/rule"
)

func main() {
	// TODO(aylei): compose according to config
	r, err := rule.NewPromReader("http://localhost:9091")
	if err != nil {
		panic(err)
	}
	ruleCli := rule.NewClient(r, &rule.NoopWriter{})

	g := api.New(ruleCli)
	g.Run(":10086")
}
