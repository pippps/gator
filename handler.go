package main

import (
	"fmt"

	"github.com/pippps/gator/internal/config"
)

type state struct {
	cfg *config.Config
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	commandMap map[string]func(*state, command) error
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return fmt.Errorf("must contain an username")
	}
	if len(cmd.arguments) > 1 {
		return fmt.Errorf("can't contain more than one arg for username")
	}

	err := config.SetUser(*s.cfg, cmd.arguments[0])
	if err != nil {
		return err
	}

	fmt.Println("User as been set")
	return nil
}

func (c *commands) run(s *state, cmd command) error {
	if err := c.commandMap[cmd.name](s, cmd); err != nil {
		return err
	}
	return nil
}
func (c *commands) register(name string, f func(*state, command) error) {
	c.commandMap[name] = f
}
