package db

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestStoreVideoFlow(t *testing.T) {
	ctx := context.Background()

	store, err := Open(filepath.Join(t.TempDir(), "feeds.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}

	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	if err := store.UpsertFeed(
		ctx,
		"Test Channel",
		"UC123",
		ChannelFeedURL("UC123"),
		"https://example.com/icon.jpg",
	); err != nil {
		t.Fatalf("upsert feed: %v", err)
	}

	feeds, err := store.ListFeeds(ctx)
	if err != nil {
		t.Fatalf("list feeds: %v", err)
	}
	if len(feeds) != 1 {
		t.Fatalf("expected 1 feed, got %d", len(feeds))
	}

	publishedAt := time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC)
	if err := store.UpsertVideo(ctx, Video{
		FeedID:      feeds[0].ID,
		VideoID:     "video-1",
		Title:       "A searchable title",
		ChannelName: "Test Channel",
		PublishedAt: publishedAt,
	}); err != nil {
		t.Fatalf("upsert video: %v", err)
	}

	videos, total, err := store.ListVideos(ctx, 1, 30, "UC123", "searchable")
	if err != nil {
		t.Fatalf("list videos: %v", err)
	}
	if total != 1 || len(videos) != 1 {
		t.Fatalf("expected 1 video and total 1, got %d and %d", len(videos), total)
	}
	if videos[0].ChannelID != "UC123" {
		t.Fatalf("expected channel id UC123, got %q", videos[0].ChannelID)
	}

	channels, err := store.RecentChannels(ctx, 10)
	if err != nil {
		t.Fatalf("recent channels: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 recent channel, got %d", len(channels))
	}
	if channels[0].UnreadCount != 1 {
		t.Fatalf("expected unread count 1, got %d", channels[0].UnreadCount)
	}

	videoURL, err := store.MarkVideoWatchedByVideoID(ctx, "video-1")
	if err != nil {
		t.Fatalf("mark video watched by video id: %v", err)
	}
	if videoURL != WatchVideoURL("video-1") {
		t.Fatalf("unexpected video url %q", videoURL)
	}

	channels, err = store.SearchChannels(ctx, "Test", 10)
	if err != nil {
		t.Fatalf("search channels: %v", err)
	}
	if len(channels) != 1 {
		t.Fatalf("expected 1 search channel, got %d", len(channels))
	}
	if channels[0].UnreadCount != 0 {
		t.Fatalf("expected unread count 0, got %d", channels[0].UnreadCount)
	}
}
