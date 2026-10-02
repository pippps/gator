package main

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
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
		return fmt.Errorf("user not registered")
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
		ID: uuid.New(),
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
		return fmt.Errorf("problem creating the user: %v", err)

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
		return fmt.Errorf("%s command takes no argument", cmd.name)
	}
	ctx := context.Background()
	s.db.DelUsers(ctx)
	fmt.Println("users table successful reset")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	if len(cmd.arguments) != 0 {
		return fmt.Errorf("%s command takes no argument", cmd.name)
	}
	ctx := context.Background()
	users, err := s.db.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("problem getting all users: %v", err)
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
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("%s command takes one argument <time_between_reqs>", cmd.name)
	}
	timeBetweenRequests, err := time.ParseDuration(cmd.arguments[0])
	if err != nil {
		return err
	}

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		if err = scrapeFeeds(s); err != nil {
			return err
		}
	}

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 2 {
		return fmt.Errorf("addfeed <name> <url>")
	}
	ctx := context.Background()

	feedParam := database.CreateFeedParams{
		ID: uuid.New(),
		CreatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		Name: cmd.arguments[0],
		Url: sql.NullString{
			String: cmd.arguments[1],
			Valid:  true,
		},
		UserID: user.ID,
	}

	feed, err := s.db.CreateFeed(ctx, feedParam)
	if err != nil {
		return err
	}

	if _, err := s.db.GetFeedFollowsForUser(ctx, feed.Name); err != nil {
		return err
	}

	fmt.Printf("feed successfully added")
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	if len(cmd.arguments) != 0 {
		return fmt.Errorf("%s command takes no argument", cmd.name)
	}
	ctx := context.Background()
	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		username, err := s.db.GetUserFromID(ctx, feed.UserID)
		if err != nil {
			return err
		}
		fmt.Printf("  name: %v, Url: %v username: %v",
			feed.Name, feed.Url, username)
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("%s command takes one argument, <url>", cmd.name)
	}

	ctx := context.Background()

	feed, err := s.db.GetFeed(ctx, sql.NullString{
		String: cmd.arguments[0],
		Valid:  true,
	})
	if err != nil {
		return fmt.Errorf("problem retrieving feed from database : %v", err)
	}

	feedFollowParam := database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UserID: user.ID,
		FeedID: feed.ID,
	}
	feedFollowRow, err := s.db.CreateFeedFollow(ctx, feedFollowParam)
	if err != nil {
		return fmt.Errorf("problem retrieving feed_follow_row from database : %v", err)
	}
	fmt.Printf("feed name: %s\nuser: %s\n", feedFollowRow.FeedName, feedFollowRow.UserName)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 0 {
		return fmt.Errorf("%s command takes no argument", cmd.name)
	}
	ctx := context.Background()
	feedsFollowsForUser, err := s.db.GetFeedFollowsForUser(ctx, user.Name)
	if err != nil {
		return err
	}

	for _, feed := range feedsFollowsForUser {
		fmt.Printf("feed name: %s\n", feed.FeedName)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("%s command need one argument: <url>", cmd.name)
	}

	feed, err := s.db.GetFeed(context.Background(), sql.NullString{
		String: cmd.arguments[0],
		Valid:  true,
	})
	if err != nil {
		return err
	}

	deleteFeedFollowForUserParams := database.DeleteFeedFollowForUserParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}
	if err = s.db.DeleteFeedFollowForUser(context.Background(), deleteFeedFollowForUserParams); err != nil {
		return err
	}
	fmt.Printf("You no longer follow %s feed", feed.Name)

	return nil
}

func handlerBrowse(s *state, cmd command) error {
	if len(cmd.arguments) > 1 {
		return fmt.Errorf("%s command accept one optional argument set to 2 if not defined: <limit>",
			cmd.name)
	}
	limit := 2
	var err error
	if len(cmd.arguments) == 1 {
		limit, err = strconv.Atoi(cmd.arguments[0])
		if err != nil {
			return err
		}
	}

	limit32 := int32(limit)
	posts, err := s.db.GetPostsForUser(context.Background(), limit32)
	if err != nil {
		return err
	}
	for _, post := range posts {
		fmt.Printf("Title: %s\nURL: %v\nPublished at: %v\nDescription: %v\n\n",
			post.Title, post.Url.String, post.PublishedAt.Time, post.Description.String)
	}
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
