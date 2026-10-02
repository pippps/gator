package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/pippps/gator/internal/database"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "gator")

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	rssFeed := &RSSFeed{}
	if err := xml.Unmarshal(data, rssFeed); err != nil {
		return nil, err
	}
	return rssFeed, err
}

func scrapeFeeds(s *state) error {
	ctx := context.Background()
	feed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		return err
	}

	if err = s.db.MarkFeedFetched(ctx, feed.ID); err != nil {
		return err
	}

	rssFeed, err := fetchFeed(ctx, feed.Url.String)

	for _, item := range rssFeed.Channel.Item {
		pubTime, err := parseStringToTime(item.PubDate)
		if err != nil {
			return err
		}
		createPostParams := database.CreatePostParams{
			ID: uuid.New(),
			CreatedAt: sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			},
			UpdatedAt: sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			},
			Title: item.Title,
			Url: sql.NullString{
				String: item.Link,
				Valid:  true,
			},
			Description: sql.NullString{
				String: item.Description,
				Valid:  true,
			},
			PublishedAt: sql.NullTime{
				Time:  pubTime,
				Valid: true,
			},
			FeedID: feed.ID,
		}
		err = s.db.CreatePost(ctx, createPostParams)
		if err != nil {
			var pqErr *pq.Error
			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		} else {
			log.Fatal(err)
		}
		}
		fmt.Printf("item title: %s\n", item.Title)
	}

	return nil
}

func parseStringToTime(pubDate string) (time.Time, error) {

	return time.Parse(time.RFC1123Z, pubDate)
}
