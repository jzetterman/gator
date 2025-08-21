package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jzetterman/gator/internal/config"
	"github.com/jzetterman/gator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	db     *database.Queries
	config *config.Config
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error resetting users: %w", err)
	}
	fmt.Println("Users table has been reset successfully")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("username not provided")
	}

	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	}

	user, err := s.db.CreateUser(context.Background(), userParams)
	if err != nil {
		if strings.Contains(err.Error(), "violates unique constraint") {
			fmt.Printf("Username %s already exists\n", userParams.Name)
			os.Exit(1)
		}
		fmt.Println(err)
		return err
	}

	fmt.Printf("User %v created successfully\n", user.Name)

	if user.ID == uuid.Nil {
		return fmt.Errorf("user not created")
	}

	command := command{
		name: "login",
		args: []string{user.Name},
	}
	err = handlerLogin(s, command)

	fmt.Println(user)

	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("username not provided")
	}

	user, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err == sql.ErrNoRows {
		fmt.Println("user not found")
		os.Exit(1)
	}
	if err != nil {
		return err
	}

	err = s.config.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Printf("%s has been set for the username\n", cmd.args[0])

	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		loggedInUser, err := s.config.ReadConfig()
		if err != nil {
			return err
		}
		if user.Name == loggedInUser.CurrentUserName {
			fmt.Printf("%s (current)\n", user.Name)
		} else {
			fmt.Println(user.Name)
		}
	}

	return nil
}

func main() {
	args := os.Args
	cmdArgs := args[1:]
	if len(args) < 2 {
		if args[0] != "reset" {
			fmt.Println("Usage: gator <command> [args]")
			os.Exit(1)

			if len(cmdArgs) < 2 {
				fmt.Println("Usage: gator <command> [args]")
				os.Exit(1)
			}
		}
	}

	currentState := &state{
		config: &config.Config{},
	}

	cfg, err := currentState.config.ReadConfig()
	if err != nil {
		panic(err)
	}
	currentState.config = &cfg

	db, err := sql.Open("postgres", cfg.Dburl)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	currentState.db = database.New(db)

	initializedCommands := &commands{
		commandMap: make(map[string]func(*state, command) error),
	}
	initializedCommands.register("register", handlerRegister)
	initializedCommands.register("login", handlerLogin)
	initializedCommands.register("reset", handlerReset)
	initializedCommands.register("users", handlerGetUsers)
	initializedCommands.register("agg", handlerAggregate)
	initializedCommands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	initializedCommands.register("feeds", handlerGetFeeds)
	initializedCommands.register("follow", middlewareLoggedIn(handlerFollow))
	initializedCommands.register("following", middlewareLoggedIn(handlerListFeedFollows))
	initializedCommands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	err = initializedCommands.run(currentState, command{name: cmdArgs[0], args: cmdArgs[1:]})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
