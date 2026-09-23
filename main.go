package main

import (
	"fmt"
	"os"

	"github.com/pippps/gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Print(err)
	}

	s := &state{
		cfg: &cfg,
	}
	c := &commands{
		commandMap: make(map[string]func(*state, command) error),
	}

	c.register("login", handlerLogin)
	args := os.Args
	if len(args) < 2 {
		fmt.Println("not enough argument: gator <command> ")
		os.Exit(1)
	}

	cmd := command{
		name:      args[1],
		arguments: args[2:],
	}
	if err := c.run(s, cmd); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
