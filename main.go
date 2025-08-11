package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jzetterman/gator/internal/config"
)

type state struct {
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	commandMap map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	if _, ok := c.commandMap[name]; ok {
		fmt.Errorf("command %s already registered", name)
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

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("username not provided")
	}

	err := s.config.SetUser(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Printf("%s has been set for the username", cmd.args[0])

	return nil
}

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Println("Usage: gator <command> [args]")
		os.Exit(1)
	}
	cmdArgs := args[1:]
	if len(cmdArgs) < 2 {
		fmt.Println("Usage: gator <command> [args]")
		os.Exit(1)
	}

	currentState := &state{
		config: &config.Config{},
	}

	cfg, err := currentState.config.ReadConfig()
	if err != nil {
		panic(err)
	}
	currentState.config = &cfg

	initializedCommands := &commands{
		commandMap: make(map[string]func(*state, command) error),
	}
	initializedCommands.register("login", handlerLogin)
	initializedCommands.run(currentState, command{name: cmdArgs[0], args: cmdArgs[1:]})
}
