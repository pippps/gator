package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/pippps/gator/internal/config"
	"github.com/pippps/gator/internal/database"
)

type state struct {
	db  *database.Queries
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
	ctx := context.Background()

	user, err := s.db.GetUser(ctx, cmd.arguments[0])
	if user.Name != cmd.arguments[0] {
		fmt.Errorf("user not registered")
		os.Exit(1)
	}

	err = config.SetUser(*s.cfg, cmd.arguments[0])
	if err != nil {
		return err
	}

	fmt.Println("User as been set")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return fmt.Errorf("must contain an username")
	}
	if len(cmd.arguments) > 1 {
		return fmt.Errorf("can't contain more than one arg for username")
	}
	ctx := context.Background()

	userParam := database.CreateUserParams{
		ID: uuid.NullUUID{
			UUID:  uuid.New(),
			Valid: true,
		},
		CreatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		Name: cmd.arguments[0],
	}

	user, err := s.db.CreateUser(ctx, userParam)
	if err != nil {
		fmt.Errorf("problem creating the user")
		os.Exit(1)
	}

	err = config.SetUser(*s.cfg, user.Name)
	if err != nil {
		return err
	}
	fmt.Println("user as been created")
	fmt.Printf("id: %v, createdAt: %v, Name: %s", user.ID, user.CreatedAt, user.Name)
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
