package main

import (
	"database/sql"

	_ "github.com/lib/pq"
	"github.com/pippps/gator/internal/database"
)

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

	c := &commands{
		commandMap: make(map[string]func(*state, command) error),
	}
	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		fmt.Errorf("error opening the database: %v", err)
	}
	dbQueries := database.New(db)
	s := &state{
		db:  dbQueries,
		cfg: &cfg,
	}

	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handlerReset)
	c.register("users", handlerUsers)
	c.register("agg", handlerAgg)
	c.register("addfeed", handlerAddFeed)

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
