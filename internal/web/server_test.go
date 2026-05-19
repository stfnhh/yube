package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tubehive/internal/db"
	"tubehive/internal/feed"
)

func TestSettingsPageAndUpdate(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	if err := os.Chdir(filepath.Join(wd, "../..")); err != nil {
		t.Fatalf("change to project root: %v", err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	}()

	ctx := context.Background()
	store, err := db.Open(filepath.Join(t.TempDir(), "feeds.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := store.UpsertFeed(
		ctx,
		"Sidebar Channel",
		"UC-sidebar",
		db.ChannelFeedURL("UC-sidebar"),
		"",
	); err != nil {
		t.Fatalf("upsert feed: %v", err)
	}

	server := New(
		store,
		feed.NewRefresher(store),
	).Routes()

	getRecorder := httptest.NewRecorder()
	server.ServeHTTP(
		getRecorder,
		httptest.NewRequest(
			http.MethodGet,
			"/settings",
			nil,
		),
	)
	if getRecorder.Code != http.StatusOK {
		t.Fatalf("expected settings status 200, got %d", getRecorder.Code)
	}
	if !strings.Contains(
		getRecorder.Body.String(),
		`href="/channels/UC-sidebar"`,
	) {
		t.Fatalf("expected settings page to include sidebar subscriptions")
	}
	if !strings.Contains(
		getRecorder.Body.String(),
		`class="nav-item active" href="/settings"`,
	) {
		t.Fatalf("expected settings nav item to be active")
	}

	form := url.Values{}
	form.Set("refresh_interval_minutes", "30")
	form.Set("video_retention_days", "180")
	form.Set("max_videos_per_channel", "500")

	postRecorder := httptest.NewRecorder()
	postRequest := httptest.NewRequest(
		http.MethodPost,
		"/settings",
		strings.NewReader(form.Encode()),
	)
	postRequest.Header.Set(
		"Content-Type",
		"application/x-www-form-urlencoded",
	)

	server.ServeHTTP(postRecorder, postRequest)
	if postRecorder.Code != http.StatusSeeOther {
		t.Fatalf("expected settings redirect, got %d", postRecorder.Code)
	}

	settings, err := store.GetSettings(ctx)
	if err != nil {
		t.Fatalf("get settings: %v", err)
	}
	if settings.RefreshIntervalMin != 30 ||
		settings.VideoRetentionDays != 180 ||
		settings.MaxVideosPerChannel != 500 {
		t.Fatalf("settings were not saved: %+v", settings)
	}
}

func TestPageURLPreservesSearchQuery(t *testing.T) {
	got := pageURL(
		"/search",
		2,
		"portable soup",
	)

	if got != "/search?page=2&q=portable+soup" &&
		got != "/search?q=portable+soup&page=2" {
		t.Fatalf("unexpected search page URL %q", got)
	}
}

func TestActiveNavSelectsChannelsForChannelViews(t *testing.T) {
	if got := activeNav("/channels/UC123"); got != "channels" {
		t.Fatalf("expected channels active nav, got %q", got)
	}
}

func TestIndexEmptyState(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	if err := os.Chdir(filepath.Join(wd, "../..")); err != nil {
		t.Fatalf("change to project root: %v", err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	}()

	ctx := context.Background()
	store, err := db.Open(filepath.Join(t.TempDir(), "feeds.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	server := New(
		store,
		feed.NewRefresher(store),
	).Routes()

	recorder := httptest.NewRecorder()
	server.ServeHTTP(
		recorder,
		httptest.NewRequest(
			http.MethodGet,
			"/",
			nil,
		),
	)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected index status 200, got %d", recorder.Code)
	}
	if !strings.Contains(
		recorder.Body.String(),
		`class="empty-state"`,
	) {
		t.Fatalf("expected index empty state")
	}
	if !strings.Contains(
		recorder.Body.String(),
		`class="primary-btn empty-action"`,
	) {
		t.Fatalf("expected styled empty state action")
	}
	if strings.Contains(
		recorder.Body.String(),
		`2 minutes ago`,
	) {
		t.Fatalf("expected last updated copy to use refresh metadata")
	}
}

func TestChannelsPage(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	if err := os.Chdir(filepath.Join(wd, "../..")); err != nil {
		t.Fatalf("change to project root: %v", err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	}()

	ctx := context.Background()
	store, err := db.Open(filepath.Join(t.TempDir(), "feeds.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := store.UpsertFeed(
		ctx,
		"Channel Page",
		"UC-channel-page",
		db.ChannelFeedURL("UC-channel-page"),
		"",
	); err != nil {
		t.Fatalf("upsert feed: %v", err)
	}
	feeds, err := store.ListFeeds(ctx)
	if err != nil {
		t.Fatalf("list feeds: %v", err)
	}
	if err := store.UpsertVideo(ctx, db.Video{
		FeedID:      feeds[0].ID,
		VideoID:     "channel-page-video",
		Title:       "Channel page video",
		ChannelName: "Channel Page",
		PublishedAt: time.Now(),
	}); err != nil {
		t.Fatalf("upsert video: %v", err)
	}

	server := New(
		store,
		feed.NewRefresher(store),
	).Routes()

	channelsRecorder := httptest.NewRecorder()
	server.ServeHTTP(
		channelsRecorder,
		httptest.NewRequest(
			http.MethodGet,
			"/channels",
			nil,
		),
	)
	if channelsRecorder.Code != http.StatusOK {
		t.Fatalf("expected channels status 200, got %d", channelsRecorder.Code)
	}
	if !strings.Contains(
		channelsRecorder.Body.String(),
		`href="/channels/UC-channel-page"`,
	) {
		t.Fatalf("expected channel page row to link to channel view")
	}
	if !strings.Contains(
		channelsRecorder.Body.String(),
		`<td class="count-column">1</td>`,
	) {
		t.Fatalf("expected channel page row to show unwatched video count")
	}
	if !strings.Contains(
		channelsRecorder.Body.String(),
		`data-delete-url="/channels/UC-channel-page"`,
	) {
		t.Fatalf("expected channel page row to include unsubscribe action")
	}
	if !strings.Contains(
		channelsRecorder.Body.String(),
		`class="nav-item active" href="/channels"`,
	) {
		t.Fatalf("expected channels nav item to be active")
	}

	unsubscribeRecorder := httptest.NewRecorder()
	server.ServeHTTP(
		unsubscribeRecorder,
		httptest.NewRequest(
			http.MethodDelete,
			"/channels/UC-channel-page",
			nil,
		),
	)
	if unsubscribeRecorder.Code != http.StatusSeeOther {
		t.Fatalf("expected unsubscribe redirect, got %d", unsubscribeRecorder.Code)
	}

	feeds, err = store.ListFeeds(ctx)
	if err != nil {
		t.Fatalf("list feeds after unsubscribe: %v", err)
	}
	if len(feeds) != 0 {
		t.Fatalf("expected channel to be unsubscribed, got %d feeds", len(feeds))
	}
}

func TestChannelViewUsesChannelTitle(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	if err := os.Chdir(filepath.Join(wd, "../..")); err != nil {
		t.Fatalf("change to project root: %v", err)
	}
	defer func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	}()

	ctx := context.Background()
	store, err := db.Open(filepath.Join(t.TempDir(), "feeds.db"))
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	if err := store.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := store.UpsertFeed(
		ctx,
		"Channel Heading",
		"UC-heading",
		db.ChannelFeedURL("UC-heading"),
		"",
	); err != nil {
		t.Fatalf("upsert feed: %v", err)
	}

	server := New(
		store,
		feed.NewRefresher(store),
	).Routes()

	recorder := httptest.NewRecorder()
	server.ServeHTTP(
		recorder,
		httptest.NewRequest(
			http.MethodGet,
			"/channels/UC-heading",
			nil,
		),
	)
	if recorder.Code != http.StatusOK {
		t.Fatalf("expected channel view status 200, got %d", recorder.Code)
	}
	if !strings.Contains(
		recorder.Body.String(),
		"<h2>Channel Heading</h2>",
	) {
		t.Fatalf("expected channel title heading")
	}
}
