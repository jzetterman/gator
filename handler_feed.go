package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jzetterman/gator/internal/database"
	_ "github.com/lib/pq"
)

func handlerAggregate(s *state, cmd command) error {
	if len(cmd.args) < 1 || len(cmd.args) > 2 {
		return fmt.Errorf("usage: %v <time_between_reqs>", cmd.name)
	}

	duration, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}
	fmt.Printf("Collecting feeds every %v\n", duration)

	ticker := time.NewTicker(duration)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func scrapeFeeds(s *state) {
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Println("error fetching next feed: ", err)
		return
	}
	fmt.Printf("Fetching feed %v\n", nextFeed.Name)
	fmt.Println("====================================")
	scrapeFeed(s.db, nextFeed)
}

func scrapeFeed(db *database.Queries, feed database.Feed) {
	currentFeed, err := db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		log.Println("error marking feed as fetched: ", err)
		return
	}
	feedData, err := fetchFeed(context.Background(), currentFeed.Url)
	if err != nil {
		log.Println("error fetching feed by URL: ", err)
		return
	}
	for _, item := range feedData.Channel.Item {
		fmt.Printf("Scraping %v\n", item.Title)
		var pubDate sql.NullTime
		parsedDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			log.Println("error parsing pubDate: ", err)
			pubDate = sql.NullTime{Valid: false}
		} else {
			pubDate = sql.NullTime{Time: parsedDate, Valid: true}
		}

		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			PublishedAt: pubDate,
			FeedID:      currentFeed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "violates unique constraint") {
				continue
			} else {
				log.Println("error creating post: ", err)
			}
			continue
		}
	}
	fmt.Println()
	log.Printf("Feed %s collected, %v posts found", feed.Name, len(feedData.Channel.Item))
}

func handlerGetPosts(s *state, cmd command, user database.User) error {
	var limit int
	if len(cmd.args) == 0 {
		limit = 2
	} else if len(cmd.args) == 1 {
		num, err := strconv.Atoi(cmd.args[0])
		limit = num
		if err != nil {
			return fmt.Errorf("invalid limit value: %v", err)
		}
	} else {
		return fmt.Errorf("usage: %v <integer value>", cmd.name)
	}
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("error getting posts: %v", err)
	}
	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Time.Format("Mon Jan 2"), post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("feed name or url not provided")
	}

	user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return err
	}

	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID,
	}
	feed, err := s.db.CreateFeed(context.Background(), feedParams)
	if err != nil {
		return err
	}
	_, err = setFeedFollowed(s, feedParams.Url)
	if err != nil {
		return err
	}

	fmt.Println("Feed created successfully:")
	printFeed(feed, user)
	fmt.Println()
	fmt.Println("=====================================")

	return nil
}

func handlerGetFeeds(s *state, cmd command) error {
	rawFeeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range rawFeeds {
		fmt.Println(feed.Name)
		fmt.Println(feed.Url)
		user, err := s.db.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("couldn't get user: %w", err)
		}
		printFeed(feed, user)
		fmt.Println("=====================================")
	}

	return nil
}

func printFeed(feed database.Feed, user database.User) {
	fmt.Printf("* ID:            %s\n", feed.ID)
	fmt.Printf("* Created:       %v\n", feed.CreatedAt)
	fmt.Printf("* Updated:       %v\n", feed.UpdatedAt)
	fmt.Printf("* Name:          %s\n", feed.Name)
	fmt.Printf("* URL:           %s\n", feed.Url)
	fmt.Printf("* User:          %s\n", user.Name)
}
