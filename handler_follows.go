package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jzetterman/gator/internal/database"
	_ "github.com/lib/pq"
)

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return errors.New("follow command requires exactly one argument")
	}
	feedURL := cmd.args[0]
	follow, err := setFeedFollowed(s, feedURL)
	if err != nil {
		return fmt.Errorf("error setting feed followed: %w", err)
	}

	fmt.Printf("%v\n", follow.FeedName)
	fmt.Printf("%v\n", follow.UserName)
	return nil
}

func setFeedFollowed(s *state, feedURL string) (database.CreateFeedFollowRow, error) {
	feed, err := s.db.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		return database.CreateFeedFollowRow{}, fmt.Errorf("error getting feed: %w", err)
	}

	currentUser, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return database.CreateFeedFollowRow{}, fmt.Errorf("error getting current user: %w", err)
	}

	followParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		FeedID:    feed.ID,
		UserID:    currentUser.ID,
	}
	followDetails, err := s.db.CreateFeedFollow(context.Background(), followParams)
	if err != nil {
		return database.CreateFeedFollowRow{}, fmt.Errorf("error creating feed follow: %w", err)
	}
	return followDetails, nil
}

func handlerListFeedFollows(s *state, cmd command, user database.User) error {
	currentUser, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return fmt.Errorf("error getting current user: %w", err)
	}
	feeds, err := s.db.GetFeedFollowsForUser(context.Background(), currentUser.ID)
	if err != nil {
		return fmt.Errorf("error getting feed follows: %w", err)
	}
	for _, feed := range feeds {
		fmt.Printf("%v\n", feed.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("usage: %s <feed_url>", cmd.name)
	}
	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("couldn't get feed: %w", err)
	}
	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("error deleting feed follow: %w", err)
	}

	fmt.Printf("%s unfollowed successfully!\n", feed.Name)
	return nil
}
