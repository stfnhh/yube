# TubeHive

Tiny server-rendered channel reader.

## Features

- Import subscriptions from OPML
- Add channels manually
- Refresh channels in a background goroutine
- Store videos in SQLite
- Render a paginated responsive grid
- Uses Pico CSS + tiny custom CSS

## Run

```sh
go mod tidy
go run ./cmd/server
```

Open http://localhost:8080.

## Manual channel subscription URL format

```text
https://<subscription-host>/feeds/videos.xml?channel_id=CHANNEL_ID
```

## Environment variables

```sh
ADDR=:8080
DB_PATH=tubehive.db
```

