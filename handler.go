package main

import (
	"context"
	"database/sql"
	"fmt"
	"html"
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

func handlerReset(s *state, cmd command) error {
	if len(cmd.arguments) != 0 {
		fmt.Errorf("too many arguments")
		os.Exit(2)
	}
	ctx := context.Background()
	s.db.DelUsers(ctx)
	fmt.Println("users table successful reset")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	if len(cmd.arguments) != 0 {
		fmt.Errorf("too many arguments")
		os.Exit(3)
	}
	ctx := context.Background()
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		fmt.Errorf("problem getting all users: %v", err)
		os.Exit(1)
	}
	for _, user := range users {
		fmt.Printf("* %v", user.Name)
		if s.cfg.CurrentUserName == user.Name {
			fmt.Printf(" (current)")
		}
		fmt.Println()
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	feedURL := "https://www.wagslane.dev/index.xml"
	ctx := context.Background()
	feed, err := fetchFeed(ctx, feedURL)
	if err != nil {
		fmt.Errorf("problem fetching the data: %v", err)
		os.Exit(5)
	}
	scapedString :=
		fmt.Sprintf("Title: %v \nLink: %v\nDescription %v\nItem:\n Title: %v \n Link: %v\n Description: %v\n",
			feed.Channel.Title, feed.Channel.Link, feed.Channel.Description,
			feed.Channel.Item[0].Title, feed.Channel.Item[0].Link, feed.Channel.Item[0].Description)
	unscapedString := html.UnescapeString(scapedString)
	fmt.Println(unscapedString)
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
