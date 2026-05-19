package feed

import "testing"

func TestExtractChannelIDFromPage(t *testing.T) {
	body := `<html><head><meta itemprop="channelId" content="UC12345678901234567890"></head></html>`

	got := extractChannelIDFromPage(body)
	if got != "UC12345678901234567890" {
		t.Fatalf("expected channel id from page, got %q", got)
	}
}

func TestExtractChannelTitleFromPage(t *testing.T) {
	body := `<html><head><meta property="og:title" content="Doug Ormsby - ` + brandName() + `"></head></html>`

	got := extractChannelTitleFromPage(body)
	if got != "Doug Ormsby" {
		t.Fatalf("expected channel title from page, got %q", got)
	}
}

func TestChannelPageURLExtractsDirectChannelID(t *testing.T) {
	rawURL := "https://" + videoHost() + "/channel/UC12345678901234567890"

	pageURL, channelID, err := channelPageURL(rawURL)
	if err != nil {
		t.Fatalf("channel page URL: %v", err)
	}
	if pageURL != "" {
		t.Fatalf("expected no page URL for direct channel id, got %q", pageURL)
	}
	if channelID != "UC12345678901234567890" {
		t.Fatalf("expected direct channel id, got %q", channelID)
	}
}

func TestChannelPageURLAllowsHandleURL(t *testing.T) {
	rawURL := "https://" + videoHost() + "/@example"

	pageURL, channelID, err := channelPageURL(rawURL)
	if err != nil {
		t.Fatalf("channel page URL: %v", err)
	}
	if pageURL == "" {
		t.Fatalf("expected handle URL to require page fetch")
	}
	if channelID != "" {
		t.Fatalf("expected no channel id before page fetch, got %q", channelID)
	}
}
