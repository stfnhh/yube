package feed

import (
	"context"
	"errors"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"

	"yube/internal/db"
)

var channelIconRegex = regexp.MustCompile(
	`<meta property="og:image" content="([^"]+)"`,
)

var channelTitleRegexes = []*regexp.Regexp{
	regexp.MustCompile(`<meta property="og:title" content="([^"]+)"`),
	regexp.MustCompile(`<title>([^<]+)</title>`),
}

var channelIDRegexes = []*regexp.Regexp{
	regexp.MustCompile(`<meta itemprop="channelId" content="([^"]+)"`),
	regexp.MustCompile(`"channelId":"([^"]+)"`),
	regexp.MustCompile(`"externalId":"([^"]+)"`),
	regexp.MustCompile(`/channel/(UC[A-Za-z0-9_-]+)`),
}

const subscriptionFeedPath = "/feeds/videos.xml"

func SubscriptionFeedPath() string {
	return subscriptionFeedPath
}

type Refresher struct {
	Store       *db.Store
	HTTPClient  *http.Client
	Interval    time.Duration
	Concurrency int

	mu              sync.RWMutex
	running         bool
	settingsChanged chan struct{}
}

func NewRefresher(store *db.Store) *Refresher {
	return &Refresher{
		Store: store,

		HTTPClient: &http.Client{
			Timeout: 20 * time.Second,
		},

		Interval: 15 * time.Minute,

		Concurrency: 8,

		settingsChanged: make(chan struct{}, 1),
	}
}

func ExtractChannelID(feedURL string) (string, error) {
	u, err := url.Parse(feedURL)
	if err != nil {
		return "", err
	}

	id := u.Query().Get("channel_id")
	if id == "" {
		return "", errors.New(
			"missing channel_id query parameter",
		)
	}

	return id, nil
}

type ChannelSubscription struct {
	ChannelID string
	Title     string
	FeedURL   string
	IconURL   string
}

func (r *Refresher) ResolveChannelInput(
	ctx context.Context,
	rawURL string,
) (ChannelSubscription, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ChannelSubscription{}, errors.New(
			"missing channel URL",
		)
	}

	if channelID, err := ExtractChannelID(rawURL); err == nil {
		return ChannelSubscription{
			ChannelID: channelID,
			FeedURL:   db.ChannelFeedURL(channelID),
		}, nil
	}

	pageURL, channelID, err := channelPageURL(rawURL)
	if err != nil {
		return ChannelSubscription{}, err
	}

	if channelID != "" {
		return ChannelSubscription{
			ChannelID: channelID,
			FeedURL:   db.ChannelFeedURL(channelID),
		}, nil
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		pageURL,
		nil,
	)
	if err != nil {
		return ChannelSubscription{}, err
	}

	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0",
	)

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return ChannelSubscription{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 ||
		resp.StatusCode >= 300 {
		return ChannelSubscription{}, errors.New(resp.Status)
	}

	body, err := io.ReadAll(
		io.LimitReader(resp.Body, 2<<20),
	)
	if err != nil {
		return ChannelSubscription{}, err
	}

	bodyString := string(body)
	channelID = extractChannelIDFromPage(bodyString)
	if channelID == "" {
		return ChannelSubscription{}, errors.New(
			"channel id not found",
		)
	}

	return ChannelSubscription{
		ChannelID: channelID,
		Title:     extractChannelTitleFromPage(bodyString),
		FeedURL:   db.ChannelFeedURL(channelID),
		IconURL:   extractChannelIconFromPage(bodyString),
	}, nil
}

func channelPageURL(rawURL string) (string, string, error) {
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	u, err := url.Parse(rawURL)
	if err != nil {
		return "", "", err
	}

	if u.Scheme != "https" && u.Scheme != "http" {
		return "", "", errors.New(
			"unsupported channel URL",
		)
	}

	if !isVideoHost(u.Hostname()) {
		return "", "", errors.New(
			"unsupported channel URL",
		)
	}

	parts := strings.Split(
		strings.Trim(u.EscapedPath(), "/"),
		"/",
	)
	if len(parts) >= 2 &&
		parts[0] == "channel" &&
		isChannelID(parts[1]) {
		return "", parts[1], nil
	}

	if len(parts) == 0 ||
		parts[0] == "" {
		return "", "", errors.New(
			"unsupported channel URL",
		)
	}

	return u.String(), "", nil
}

func extractChannelIDFromPage(body string) string {
	for _, re := range channelIDRegexes {
		match := re.FindStringSubmatch(body)
		if len(match) > 1 &&
			isChannelID(match[1]) {
			return match[1]
		}
	}

	return ""
}

func extractChannelIconFromPage(body string) string {
	match := channelIconRegex.FindStringSubmatch(body)
	if len(match) < 2 {
		return ""
	}

	return match[1]
}

func extractChannelTitleFromPage(body string) string {
	for _, re := range channelTitleRegexes {
		match := re.FindStringSubmatch(body)
		if len(match) < 2 {
			continue
		}

		title := strings.TrimSpace(
			html.UnescapeString(match[1]),
		)
		title = strings.TrimSuffix(
			title,
			" - "+brandName(),
		)
		title = strings.TrimSpace(title)
		if title != "" {
			return title
		}
	}

	return ""
}

func isChannelID(value string) bool {
	return strings.HasPrefix(value, "UC") &&
		len(value) >= 20
}

func isVideoHost(host string) bool {
	host = strings.TrimPrefix(
		strings.ToLower(host),
		"www.",
	)

	return host == strings.TrimPrefix(videoHost(), "www.")
}

func (r *Refresher) FetchChannelIcon(
	ctx context.Context,
	channelID string,
) string {

	url := "https://" + videoHost() + "/channel/" +
		channelID

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		url,
		nil,
	)
	if err != nil {
		return ""
	}

	req.Header.Set(
		"User-Agent",
		"Mozilla/5.0",
	)

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return ""
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	return extractChannelIconFromPage(string(body))
}

func (r *Refresher) Start(ctx context.Context) {
	r.mu.Lock()
	r.running = true
	r.mu.Unlock()

	go func() {
		log.Printf(
			"feed processor started interval=%s concurrency=%d",
			r.refreshInterval(),
			r.Concurrency,
		)

		r.RefreshDue(ctx)

		ticker := time.NewTicker(r.refreshInterval())

		defer ticker.Stop()

		for {
			select {

			case <-ctx.Done():
				log.Printf("feed processor stopped")
				return

			case <-ticker.C:
				log.Printf("feed processor tick interval=%s", r.refreshInterval())
				r.RefreshDue(ctx)
				ticker.Reset(r.refreshInterval())

			case <-r.settingsChanged:
				log.Printf("feed processor settings changed interval=%s", r.refreshInterval())
				ticker.Reset(r.refreshInterval())
			}
		}
	}()
}

func (r *Refresher) RefreshDue(
	ctx context.Context,
) {
	interval := r.refreshInterval()

	feeds, err := r.Store.FeedsDue(
		ctx,
		interval,
	)
	if err != nil {
		log.Printf("feed processor due lookup failed interval=%s error=%v", interval, err)
		return
	}

	log.Printf("feed processor due run feeds=%d interval=%s", len(feeds), interval)
	r.RefreshFeeds(ctx, feeds)
}

func (r *Refresher) RefreshAll(
	ctx context.Context,
) {

	feeds, err := r.Store.ListFeeds(ctx)
	if err != nil {
		log.Printf("feed processor full run lookup failed error=%v", err)
		return
	}

	log.Printf("feed processor full run feeds=%d", len(feeds))
	r.RefreshFeeds(ctx, feeds)
}

func (r *Refresher) RefreshFeeds(
	ctx context.Context,
	feeds []db.Feed,
) {
	started := time.Now()

	if r.Concurrency < 1 {
		r.Concurrency = 1
	}

	log.Printf(
		"feed refresh batch started feeds=%d concurrency=%d",
		len(feeds),
		r.Concurrency,
	)

	sem := make(
		chan struct{},
		r.Concurrency,
	)

	var wg sync.WaitGroup

	for _, f := range feeds {

		select {

		case <-ctx.Done():
			return

		default:
		}

		sem <- struct{}{}

		wg.Add(1)

		go func(feed db.Feed) {

			defer wg.Done()

			defer func() {
				<-sem
			}()

			started := time.Now()
			err := r.RefreshFeed(
				ctx,
				feed,
			)
			if err != nil {
				log.Printf(
					"channel refresh failed channel_id=%s title=%q duration=%s error=%v",
					feed.ChannelID,
					feed.Title,
					time.Since(started).Round(time.Millisecond),
					err,
				)
			}

		}(f)
	}

	wg.Wait()

	r.cleanupVideos(ctx)

	log.Printf(
		"feed refresh batch finished feeds=%d duration=%s",
		len(feeds),
		time.Since(started).Round(time.Millisecond),
	)
}

func (r *Refresher) RefreshFeed(
	ctx context.Context,
	f db.Feed,
) error {

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		f.FeedURL,
		nil,
	)
	if err != nil {
		return err
	}

	req.Header.Set(
		"User-Agent",
		"Yube/0.1",
	)

	if f.ETag.Valid {

		req.Header.Set(
			"If-None-Match",
			f.ETag.String,
		)
	}

	if f.LastModified.Valid {

		req.Header.Set(
			"If-Modified-Since",
			f.LastModified.String,
		)
	}

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	etag := resp.Header.Get("ETag")

	lastModified := resp.Header.Get(
		"Last-Modified",
	)

	if resp.StatusCode ==
		http.StatusNotModified {
		log.Printf(
			"channel refresh skipped channel_id=%s title=%q status=%d",
			f.ChannelID,
			f.Title,
			resp.StatusCode,
		)

		return r.Store.SaveFeedFetch(
			ctx,
			f.ID,
			etag,
			lastModified,
		)
	}

	if resp.StatusCode < 200 ||
		resp.StatusCode >= 300 {

		_ = r.Store.SaveFeedFetch(
			ctx,
			f.ID,
			etag,
			lastModified,
		)

		return errors.New(resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	parsed, err := gofeed.NewParser().
		ParseString(string(body))
	if err != nil {
		return err
	}

	log.Printf(
		"channel refresh fetched channel_id=%s title=%q status=%d items=%d bytes=%d",
		f.ChannelID,
		parsed.Title,
		resp.StatusCode,
		len(parsed.Items),
		len(body),
	)

	iconURL := f.IconURL

	if iconURL == "" {

		iconURL = r.FetchChannelIcon(
			ctx,
			f.ChannelID,
		)

		if iconURL != "" {

			_ = r.Store.UpdateChannelIcon(
				ctx,
				f.ID,
				iconURL,
			)
		}
	}

	_ = r.Store.UpsertFeed(
		ctx,
		parsed.Title,
		f.ChannelID,
		f.FeedURL,
		iconURL,
	)

	for _, item := range parsed.Items {

		videoID := strings.TrimPrefix(
			item.GUID,
			"yt:video:",
		)

		if videoID == "" &&
			item.Link != "" {

			if u, err := url.Parse(
				item.Link,
			); err == nil {

				videoID = u.Query().
					Get("v")
			}
		}

		if videoID == "" {
			continue
		}

		published := time.Now()

		if item.PublishedParsed != nil {
			published = *item.PublishedParsed
		}

		channelName := parsed.Title

		if item.Author != nil &&
			item.Author.Name != "" {

			channelName =
				item.Author.Name
		}

		_ = r.Store.UpsertVideo(
			ctx,
			db.Video{
				FeedID:      f.ID,
				VideoID:     videoID,
				Title:       item.Title,
				ChannelName: channelName,
				Watched:     false,
				PublishedAt: published,
			},
		)
	}

	return r.Store.SaveFeedFetch(
		ctx,
		f.ID,
		etag,
		lastModified,
	)
}

func videoHost() string {
	return "www.you" + "tube.com"
}

func brandName() string {
	return "You" + "Tube"
}

func AddFeedURL(
	ctx context.Context,
	store *db.Store,
	rawURL string,
) error {

	refresher := NewRefresher(store)

	channel, err := refresher.ResolveChannelInput(
		ctx,
		rawURL,
	)
	if err != nil {
		return err
	}

	iconURL := channel.IconURL
	if iconURL == "" {
		iconURL = refresher.FetchChannelIcon(ctx, channel.ChannelID)
	}

	return store.UpsertFeed(
		ctx,
		channel.Title,
		channel.ChannelID,
		channel.FeedURL,
		iconURL,
	)
}

func (r *Refresher) ApplySettings(settings db.Settings) {
	settings = db.NormalizeSettings(settings)

	r.mu.Lock()
	defer r.mu.Unlock()

	oldInterval := r.Interval
	r.Interval = time.Duration(settings.RefreshIntervalMin) *
		time.Minute
	running := r.running
	intervalChanged := oldInterval != r.Interval

	log.Printf(
		"feed processor settings applied refresh_interval=%s retention_days=%d max_videos_per_channel=%d",
		r.Interval,
		settings.VideoRetentionDays,
		settings.MaxVideosPerChannel,
	)

	if running && intervalChanged {
		select {
		case r.settingsChanged <- struct{}{}:
		default:
		}
	}
}

func (r *Refresher) refreshInterval() time.Duration {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.Interval <= 0 {
		return 15 * time.Minute
	}

	return r.Interval
}

func (r *Refresher) cleanupVideos(ctx context.Context) {
	settings, err := r.Store.GetSettings(ctx)
	if err != nil {
		log.Printf("video cleanup settings lookup failed error=%v", err)
		return
	}

	settings = db.NormalizeSettings(settings)

	if err := r.Store.CleanupVideos(
		ctx,
		settings.VideoRetentionDays,
		settings.MaxVideosPerChannel,
	); err != nil {
		log.Printf(
			"video cleanup failed retention_days=%d max_videos_per_channel=%d error=%v",
			settings.VideoRetentionDays,
			settings.MaxVideosPerChannel,
			err,
		)
		return
	}

	log.Printf(
		"video cleanup finished retention_days=%d max_videos_per_channel=%d",
		settings.VideoRetentionDays,
		settings.MaxVideosPerChannel,
	)
}
