# Gator 

Gator is a command-line RSS feed aggregator written in Go. It lets you register users, add RSS feeds, follow feeds added by others, continuously fetch new posts in the background, and browse the latest posts right from your terminal.

## Prerequisites

You'll need these installed to run gator:

- **Go** (1.22 or newer): https://go.dev/doc/install
- **PostgreSQL** (15 or newer): https://www.postgresql.org/download/

Make sure Postgres is running and create a database for gator:

```bash
psql postgres
CREATE DATABASE gator;
```

## Installation

Install the `gator` CLI with `go install`:

```bash
go install github.com/<your-github-username>/gator@latest
```

This compiles the program and puts the `gator` binary in your `$GOPATH/bin` (usually `~/go/bin`). Make sure that directory is on your `PATH` so you can run `gator` from anywhere.

> Go programs are statically compiled, so once installed, `gator` runs without needing the Go toolchain. Use `go run .` only during development.

## Configuration

Gator reads its settings from a JSON file in your home directory: `~/.gatorconfig.json`. Create it with your Postgres connection string:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable"
}
```

Replace `username` and `password` with your Postgres credentials. Gator adds a `current_user_name` field to this file automatically when you register or log in.

### Database migrations

The schema lives in `sql/schema`. Run the migrations with [goose](https://github.com/pressly/goose):

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
cd sql/schema
goose postgres "postgres://username:password@localhost:5432/gator" up
```

## Usage

```bash
gator <command> [args...]
```

### Users

| Command | Description |
|---|---|
| `gator register <name>` | Create a new user and log in as them |
| `gator login <name>` | Switch to an existing user |
| `gator users` | List all users (the current one is marked) |
| `gator reset` | Delete all users and data (useful for testing) |

### Feeds

| Command | Description |
|---|---|
| `gator addfeed <name> <url>` | Add a feed and automatically follow it |
| `gator feeds` | List all feeds in the database |
| `gator follow <url>` | Follow an existing feed |
| `gator following` | List the feeds the current user follows |
| `gator unfollow <url>` | Stop following a feed |

### Aggregating and browsing

| Command | Description |
|---|---|
| `gator agg <time_between_reqs>` | Fetch feeds continuously, e.g. `gator agg 1m`. Runs until you stop it with Ctrl+C |
| `gator browse [limit]` | Show the latest posts from feeds you follow (default 2) |

### Example session

```bash
gator register alice
gator addfeed "Hacker News" https://hnrss.org/newest
gator addfeed "Boot.dev Blog" https://blog.boot.dev/index.xml
gator agg 30s        # leave running in one terminal
gator browse 10      # in another terminal
```

Be kind to the servers you're fetching from: don't set the interval too low.
