package main

import (
	"errors"
	"fmt"
)

type command struct {
	name string
	args []string
}

type commands struct {
	commandMap map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	if _, ok := c.commandMap[name]; ok {
		fmt.Printf("command %s already registered\n", name)
		return
	}
	c.commandMap[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	if _, ok := c.commandMap[cmd.name]; !ok {
		return errors.New("command " + cmd.name + " not found")
	}
	return c.commandMap[cmd.name](s, cmd)
}
