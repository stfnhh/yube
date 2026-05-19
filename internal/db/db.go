package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/sqliteshim"
)

type Feed struct {
	bun.BaseModel `bun:"table:feeds,alias:f"`

	ID            int64          `bun:"id,pk,autoincrement"`
	Title         string         `bun:"title,notnull"`
	ChannelID     string         `bun:"channel_id,notnull,unique"`
	FeedURL       string         `bun:"feed_url,notnull,unique"`
	IconURL       string         `bun:"icon_url,notnull,default:''"`
	ETag          sql.NullString `bun:"etag"`
	LastModified  sql.NullString `bun:"last_modified"`
	LastCheckedAt sql.NullTime   `bun:"last_checked_at"`
	CreatedAt     time.Time      `bun:"created_at,notnull,default:CURRENT_TIMESTAMP"`
}

type RecentChannel struct {
	FeedID          int64
	Title           string
	ChannelID       string
	IconURL         string
	UnreadCount     int
	LastPublishedAt time.Time
}

type ChannelFeed struct {
	ID            int64
	Title         string
	ChannelID     string
	FeedURL       string
	IconURL       string
	UnreadCount   int
	LastCheckedAt sql.NullTime
	CreatedAt     time.Time
}

type recentChannelRow struct {
	FeedID          int64          `bun:"feed_id"`
	Title           string         `bun:"title"`
	ChannelID       string         `bun:"channel_id"`
	IconURL         string         `bun:"icon_url"`
	UnreadCount     int            `bun:"unread_count"`
	LastPublishedAt sql.NullString `bun:"last_published_at"`
}

type Video struct {
	bun.BaseModel `bun:"table:videos,alias:v"`

	ID           int64     `bun:"id,pk,autoincrement"`
	FeedID       int64     `bun:"feed_id,notnull"`
	VideoID      string    `bun:"video_id,notnull,unique"`
	Title        string    `bun:"title,notnull"`
	VideoURL     string    `bun:"video_url,notnull"`
	ThumbnailURL string    `bun:"thumbnail_url,notnull"`
	ChannelID    string    `bun:"-"`
	ChannelName  string    `bun:"channel_name,notnull"`
	Watched      bool      `bun:"watched,notnull,default:0"`
	PublishedAt  time.Time `bun:"published_at,notnull"`
	CreatedAt    time.Time `bun:"created_at,notnull,default:CURRENT_TIMESTAMP"`
}

type Settings struct {
	bun.BaseModel `bun:"table:settings,alias:s"`

	ID                  int64        `bun:"id,pk"`
	RefreshIntervalMin  int          `bun:"refresh_interval_minutes,notnull"`
	VideoRetentionDays  int          `bun:"video_retention_days,notnull"`
	MaxVideosPerChannel int          `bun:"max_videos_per_channel,notnull"`
	LastRefreshedAt     sql.NullTime `bun:"last_refreshed_at"`
	UpdatedAt           time.Time    `bun:"updated_at,notnull,default:CURRENT_TIMESTAMP"`
}

type videoRow struct {
	bun.BaseModel `bun:"table:videos,alias:v"`

	ID           int64     `bun:"id"`
	FeedID       int64     `bun:"feed_id"`
	VideoID      string    `bun:"video_id"`
	Title        string    `bun:"title"`
	VideoURL     string    `bun:"video_url"`
	ThumbnailURL string    `bun:"thumbnail_url"`
	ChannelID    string    `bun:"channel_id"`
	ChannelName  string    `bun:"channel_name"`
	Watched      bool      `bun:"watched"`
	PublishedAt  time.Time `bun:"published_at"`
	CreatedAt    time.Time `bun:"created_at"`
}

type Store struct {
	DB *bun.DB
}

const (
	videoHost     = "www.youtube.com"
	thumbnailHost = "i.ytimg.com"
)

func Open(path string) (*Store, error) {
	database, err := sql.Open(
		sqliteshim.ShimName,
		path+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)",
	)
	if err != nil {
		return nil, err
	}

	if err := database.Ping(); err != nil {
		return nil, err
	}

	return &Store{
		DB: bun.NewDB(
			database,
			sqlitedialect.New(),
		),
	}, nil
}

func (s *Store) Migrate(ctx context.Context) error {
	if _, err := s.DB.NewCreateTable().
		Model((*Feed)(nil)).
		IfNotExists().
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.DB.NewCreateTable().
		Model((*Video)(nil)).
		IfNotExists().
		ForeignKey(
			`("feed_id") REFERENCES "feeds" ("id") ON DELETE CASCADE`,
		).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.DB.NewCreateIndex().
		Model((*Video)(nil)).
		IfNotExists().
		Index("idx_videos_published_at").
		ColumnExpr("published_at DESC").
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.DB.NewCreateIndex().
		Model((*Video)(nil)).
		IfNotExists().
		Index("idx_videos_feed_id").
		Column("feed_id").
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.DB.NewCreateTable().
		Model((*Settings)(nil)).
		IfNotExists().
		Exec(ctx); err != nil {
		return err
	}

	settings := DefaultSettings()

	_, err := s.DB.NewInsert().
		Model(&settings).
		Column(
			"id",
			"refresh_interval_minutes",
			"video_retention_days",
			"max_videos_per_channel",
		).
		On("CONFLICT(id) DO NOTHING").
		Exec(ctx)

	return err
}

func (s *Store) UpsertFeed(
	ctx context.Context,
	title,
	channelID,
	feedURL,
	iconURL string,
) error {

	if title == "" {
		title = channelID
	}

	feed := &Feed{
		Title:     title,
		ChannelID: channelID,
		FeedURL:   feedURL,
		IconURL:   iconURL,
	}

	_, err := s.DB.NewInsert().
		Model(feed).
		Column(
			"title",
			"channel_id",
			"feed_url",
			"icon_url",
		).
		On("CONFLICT(channel_id) DO UPDATE").
		Set("title = EXCLUDED.title").
		Set("feed_url = EXCLUDED.feed_url").
		Set(
			"icon_url = COALESCE(NULLIF(EXCLUDED.icon_url, ''), icon_url)",
		).
		Exec(ctx)

	return err
}

func (s *Store) UpdateChannelIcon(
	ctx context.Context,
	feedID int64,
	iconURL string,
) error {

	_, err := s.DB.NewUpdate().
		Model((*Feed)(nil)).
		Set("icon_url = ?", iconURL).
		Where("id = ?", feedID).
		Exec(ctx)

	return err
}

func (s *Store) ChannelIconURL(
	ctx context.Context,
	channelID string,
) (string, error) {
	var iconURL string

	err := s.DB.NewSelect().
		Model((*Feed)(nil)).
		Column("icon_url").
		Where("channel_id = ?", channelID).
		Scan(ctx, &iconURL)

	return iconURL, err
}

func (s *Store) FeedByChannelID(
	ctx context.Context,
	channelID string,
) (Feed, error) {
	var feed Feed

	err := s.DB.NewSelect().
		Model(&feed).
		Where("channel_id = ?", channelID).
		Scan(ctx)

	return feed, err
}

func (s *Store) DeleteFeedByChannelID(
	ctx context.Context,
	channelID string,
) error {
	_, err := s.DB.NewDelete().
		Model((*Feed)(nil)).
		Where("channel_id = ?", channelID).
		Exec(ctx)

	return err
}

func (s *Store) VideoThumbnailURL(
	ctx context.Context,
	videoID string,
) (string, error) {
	var storedVideoID string

	err := s.DB.NewSelect().
		Model((*Video)(nil)).
		Column("video_id").
		Where("video_id = ?", videoID).
		Scan(ctx, &storedVideoID)
	if err != nil {
		return "", err
	}

	return ThumbnailURL(storedVideoID), nil
}

func (s *Store) ListFeeds(
	ctx context.Context,
) ([]Feed, error) {
	var feeds []Feed

	err := s.DB.NewSelect().
		Model(&feeds).
		Order("title").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return feeds, nil
}

func (s *Store) ListChannelFeeds(
	ctx context.Context,
) ([]ChannelFeed, error) {
	var feeds []ChannelFeed

	err := s.DB.NewSelect().
		Model((*Feed)(nil)).
		ColumnExpr("f.id AS id").
		ColumnExpr("f.title AS title").
		ColumnExpr("f.channel_id AS channel_id").
		ColumnExpr("f.feed_url AS feed_url").
		ColumnExpr("f.icon_url AS icon_url").
		ColumnExpr("f.last_checked_at AS last_checked_at").
		ColumnExpr("f.created_at AS created_at").
		ColumnExpr(
			"COUNT(CASE WHEN v.watched = ? THEN 1 END) AS unread_count",
			false,
		).
		Join("LEFT JOIN videos AS v ON v.feed_id = f.id").
		Group(
			"f.id",
			"f.title",
			"f.channel_id",
			"f.feed_url",
			"f.icon_url",
			"f.last_checked_at",
			"f.created_at",
		).
		Order("f.title").
		Scan(ctx, &feeds)
	if err != nil {
		return nil, err
	}

	return feeds, nil
}

func (s *Store) RecentChannels(
	ctx context.Context,
	limit int,
) ([]RecentChannel, error) {
	if limit < 1 {
		limit = 10
	}

	var rows []recentChannelRow

	err := s.DB.NewSelect().
		Model((*Feed)(nil)).
		ColumnExpr("f.id AS feed_id").
		ColumnExpr("f.title AS title").
		ColumnExpr("f.channel_id AS channel_id").
		ColumnExpr("f.icon_url AS icon_url").
		ColumnExpr(
			"COUNT(CASE WHEN v.watched = ? THEN 1 END) AS unread_count",
			false,
		).
		ColumnExpr("MAX(v.published_at) AS last_published_at").
		Join("LEFT JOIN videos AS v ON v.feed_id = f.id").
		Group("f.id", "f.title", "f.channel_id", "f.icon_url").
		OrderExpr("last_published_at DESC").
		Limit(limit).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	return recentChannelRows(rows)
}

func (s *Store) SearchChannels(
	ctx context.Context,
	search string,
	limit int,
) ([]RecentChannel, error) {

	search = strings.TrimSpace(search)
	if search == "" {
		return nil, nil
	}

	if limit < 1 {
		limit = 8
	}

	pattern := "%" + search + "%"

	var rows []recentChannelRow

	err := s.DB.NewSelect().
		Model((*Feed)(nil)).
		ColumnExpr("f.id AS feed_id").
		ColumnExpr("f.title AS title").
		ColumnExpr("f.channel_id AS channel_id").
		ColumnExpr("f.icon_url AS icon_url").
		ColumnExpr(
			"COUNT(CASE WHEN v.watched = ? THEN 1 END) AS unread_count",
			false,
		).
		ColumnExpr("MAX(v.published_at) AS last_published_at").
		Join("LEFT JOIN videos AS v ON v.feed_id = f.id").
		WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.
				WhereOr("f.title LIKE ?", pattern).
				WhereOr("f.channel_id LIKE ?", pattern)
		}).
		Group("f.id", "f.title", "f.channel_id", "f.icon_url").
		OrderExpr(
			"CASE WHEN f.title LIKE ? THEN 0 ELSE 1 END",
			search+"%",
		).
		OrderExpr("last_published_at DESC").
		Limit(limit).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	return recentChannelRows(rows)
}

func (s *Store) FeedsDue(
	ctx context.Context,
	olderThan time.Duration,
) ([]Feed, error) {

	cutoff := time.Now().Add(-olderThan)

	var feeds []Feed

	err := s.DB.NewSelect().
		Model(&feeds).
		Where("last_checked_at IS NULL OR last_checked_at < ?", cutoff).
		OrderExpr("last_checked_at IS NOT NULL").
		Order("last_checked_at").
		Scan(ctx)
	if err != nil {
		return nil, err
	}

	return feeds, nil
}

func (s *Store) LastRefreshTime(ctx context.Context) (time.Time, bool, error) {
	var lastCheckedAt sql.NullString

	err := s.DB.NewSelect().
		Model((*Feed)(nil)).
		ColumnExpr("MAX(last_checked_at)").
		Scan(ctx, &lastCheckedAt)
	if err != nil {
		return time.Time{}, false, err
	}

	lastUpdated, ok, err := parseOptionalDBTime(lastCheckedAt)
	if err != nil {
		return time.Time{}, false, err
	}

	var lastRefreshedAt sql.NullString

	err = s.DB.NewSelect().
		Model((*Settings)(nil)).
		Column("last_refreshed_at").
		Where("id = ?", 1).
		Scan(ctx, &lastRefreshedAt)
	if err != nil {
		return time.Time{}, false, err
	}

	refreshedAt, refreshedOK, err := parseOptionalDBTime(lastRefreshedAt)
	if err != nil {
		return time.Time{}, false, err
	}

	if refreshedOK &&
		(!ok || refreshedAt.After(lastUpdated)) {
		return refreshedAt, true, nil
	}

	return lastUpdated, ok, nil
}

func (s *Store) TouchRefreshTime(ctx context.Context) error {
	_, err := s.DB.NewUpdate().
		Model((*Settings)(nil)).
		Set("last_refreshed_at = CURRENT_TIMESTAMP").
		Where("id = ?", 1).
		Exec(ctx)

	return err
}

func (s *Store) SaveFeedFetch(
	ctx context.Context,
	feedID int64,
	etag,
	lastModified string,
) error {

	_, err := s.DB.NewUpdate().
		Model((*Feed)(nil)).
		SetColumn("etag", "NULLIF(?, '')", etag).
		SetColumn("last_modified", "NULLIF(?, '')", lastModified).
		SetColumn("last_checked_at", "CURRENT_TIMESTAMP").
		Where("id = ?", feedID).
		Exec(ctx)

	return err
}

func (s *Store) MarkVideoWatched(
	ctx context.Context,
	videoID int64,
) error {

	_, err := s.DB.NewUpdate().
		Model((*Video)(nil)).
		Set("watched = ?", true).
		Where("id = ?", videoID).
		Exec(ctx)

	return err
}

func (s *Store) MarkVideoWatchedByVideoID(
	ctx context.Context,
	videoID string,
) (string, error) {
	result, err := s.DB.NewUpdate().
		Model((*Video)(nil)).
		Set("watched = ?", true).
		Where("video_id = ?", videoID).
		Exec(ctx)
	if err != nil {
		return "", err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return "", err
	}

	if rowsAffected == 0 {
		return "", sql.ErrNoRows
	}

	return WatchVideoURL(videoID), nil
}

func (s *Store) UpsertVideo(
	ctx context.Context,
	v Video,
) error {
	v.VideoURL = ""
	v.ThumbnailURL = ""

	_, err := s.DB.NewInsert().
		Model(&v).
		Column(
			"feed_id",
			"video_id",
			"title",
			"video_url",
			"thumbnail_url",
			"channel_name",
			"watched",
			"published_at",
		).
		On("CONFLICT(video_id) DO UPDATE").
		Set("title = EXCLUDED.title").
		Set("video_url = EXCLUDED.video_url").
		Set("thumbnail_url = EXCLUDED.thumbnail_url").
		Set("channel_name = EXCLUDED.channel_name").
		Set("published_at = EXCLUDED.published_at").
		Exec(ctx)

	return err
}

func (s *Store) ListVideos(
	ctx context.Context,
	page,
	perPage int,
	channelID string,
	search string,
) ([]Video, int, error) {

	if page < 1 {
		page = 1
	}

	if perPage < 1 {
		perPage = 30
	}

	offset := (page - 1) * perPage
	search = strings.TrimSpace(search)

	countQuery := s.videoListQuery(
		channelID,
		search,
	)

	total, err := countQuery.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	var rows []videoRow

	err = s.videoListQuery(
		channelID,
		search,
	).
		ColumnExpr("v.id AS id").
		ColumnExpr("v.feed_id AS feed_id").
		ColumnExpr("v.video_id AS video_id").
		ColumnExpr("v.title AS title").
		ColumnExpr("f.channel_id AS channel_id").
		ColumnExpr("v.channel_name AS channel_name").
		ColumnExpr("v.watched AS watched").
		ColumnExpr("v.published_at AS published_at").
		ColumnExpr("v.created_at AS created_at").
		OrderExpr("v.published_at DESC").
		Limit(perPage).
		Offset(offset).
		Scan(ctx, &rows)
	if err != nil {
		return nil, 0, err
	}

	videos := make([]Video, 0, len(rows))
	for _, row := range rows {
		videos = append(videos, row.Video())
	}

	return videos, total, nil
}

func (s *Store) videoListQuery(
	channelID string,
	search string,
) *bun.SelectQuery {
	query := s.DB.NewSelect().
		Model((*Video)(nil)).
		Join("JOIN feeds AS f ON f.id = v.feed_id")

	if channelID != "" {
		query = query.Where("f.channel_id = ?", channelID)
	}

	if search != "" {
		pattern := "%" + search + "%"

		query = query.WhereGroup(" AND ", func(q *bun.SelectQuery) *bun.SelectQuery {
			return q.
				WhereOr("v.title LIKE ?", pattern).
				WhereOr("v.channel_name LIKE ?", pattern).
				WhereOr("f.title LIKE ?", pattern).
				WhereOr("f.channel_id LIKE ?", pattern)
		})
	}

	return query
}

func (s *Store) CleanupVideos(ctx context.Context, maxAgeDays int, maxPerFeed int) error {
	_, err := s.DB.NewDelete().
		Model((*Video)(nil)).
		Where(
			"published_at < datetime('now', ?)",
			fmt.Sprintf("-%d days", maxAgeDays),
		).
		Exec(ctx)
	if err != nil {
		return err
	}

	rankedVideos := s.DB.NewSelect().
		Model((*Video)(nil)).
		Column("id").
		ColumnExpr(
			"ROW_NUMBER() OVER (PARTITION BY feed_id ORDER BY published_at DESC) AS rn",
		)

	staleVideos := s.DB.NewSelect().
		TableExpr("(?) AS ranked_videos", rankedVideos).
		Column("id").
		Where("rn > ?", maxPerFeed)

	_, err = s.DB.NewDelete().
		Model((*Video)(nil)).
		Where("id IN (?)", staleVideos).
		Exec(ctx)

	return err
}

func (s *Store) GetSettings(ctx context.Context) (Settings, error) {
	settings := DefaultSettings()

	err := s.DB.NewSelect().
		Model(&settings).
		Where("id = ?", 1).
		Scan(ctx)
	if err != nil {
		return Settings{}, err
	}

	return settings, nil
}

func (s *Store) SaveSettings(
	ctx context.Context,
	settings Settings,
) error {
	settings.ID = 1
	settings = NormalizeSettings(settings)

	_, err := s.DB.NewInsert().
		Model(&settings).
		Column(
			"id",
			"refresh_interval_minutes",
			"video_retention_days",
			"max_videos_per_channel",
		).
		On("CONFLICT(id) DO UPDATE").
		Set("refresh_interval_minutes = EXCLUDED.refresh_interval_minutes").
		Set("video_retention_days = EXCLUDED.video_retention_days").
		Set("max_videos_per_channel = EXCLUDED.max_videos_per_channel").
		Set("updated_at = CURRENT_TIMESTAMP").
		Exec(ctx)

	return err
}

func DefaultSettings() Settings {
	return Settings{
		ID:                  1,
		RefreshIntervalMin:  15,
		VideoRetentionDays:  90,
		MaxVideosPerChannel: 250,
	}
}

func NormalizeSettings(settings Settings) Settings {
	if settings.RefreshIntervalMin < 5 {
		settings.RefreshIntervalMin = 15
	}

	if settings.VideoRetentionDays < 7 {
		settings.VideoRetentionDays = 90
	}

	if settings.MaxVideosPerChannel < 25 {
		settings.MaxVideosPerChannel = 250
	}

	return settings
}

func recentChannelRows(rows []recentChannelRow) ([]RecentChannel, error) {
	channels := make([]RecentChannel, 0, len(rows))

	for _, row := range rows {
		channel := RecentChannel{
			FeedID:      row.FeedID,
			Title:       row.Title,
			ChannelID:   row.ChannelID,
			IconURL:     row.IconURL,
			UnreadCount: row.UnreadCount,
		}

		if row.LastPublishedAt.Valid {
			publishedAt, err := parseDBTime(row.LastPublishedAt.String)
			if err != nil {
				return nil, err
			}

			channel.LastPublishedAt = publishedAt
		}

		channels = append(channels, channel)
	}

	return channels, nil
}

func parseDBTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}

	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999Z07:00",
		"2006-01-02 15:04:05.999999999 -0700 MST",
		"2006-01-02 15:04:05 -0700 MST",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	}

	var parseErr error
	for _, format := range formats {
		parsed, err := time.Parse(format, value)
		if err == nil {
			return parsed, nil
		}

		parseErr = err
	}

	return time.Time{}, parseErr
}

func parseOptionalDBTime(value sql.NullString) (time.Time, bool, error) {
	if !value.Valid {
		return time.Time{}, false, nil
	}

	parsed, err := parseDBTime(value.String)
	if err != nil {
		return time.Time{}, false, err
	}

	return parsed, !parsed.IsZero(), nil
}

func (row videoRow) Video() Video {
	return Video{
		ID:           row.ID,
		FeedID:       row.FeedID,
		VideoID:      row.VideoID,
		Title:        row.Title,
		VideoURL:     WatchVideoURL(row.VideoID),
		ThumbnailURL: ThumbnailURL(row.VideoID),
		ChannelID:    row.ChannelID,
		ChannelName:  row.ChannelName,
		Watched:      row.Watched,
		PublishedAt:  row.PublishedAt,
		CreatedAt:    row.CreatedAt,
	}
}

func WatchVideoURL(videoID string) string {
	return fmt.Sprintf(
		"https://%s/watch?v=%s",
		videoHost,
		videoID,
	)
}

func ThumbnailURL(videoID string) string {
	return fmt.Sprintf(
		"https://%s/vi/%s/hqdefault.jpg",
		thumbnailHost,
		videoID,
	)
}

func ChannelFeedURL(channelID string) string {
	return fmt.Sprintf(
		"https://%s/feeds/videos.xml?channel_id=%s",
		videoHost,
		channelID,
	)
}
