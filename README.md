# Yubè

Tiny server-rendered YouTube feed reader.

## Features

- Import OPML files
- Add YouTube Atom feeds manually
- Fetch feeds in a background goroutine
- Store videos in SQLite
- Render a paginated responsive grid
- Uses Pico CSS + tiny custom CSS

## Run

```sh
go mod tidy
go run ./cmd/server
```

Open http://localhost:8080.

## Manual feed URL format

```text
https://www.youtube.com/feeds/videos.xml?channel_id=CHANNEL_ID
```

## Environment variables

```sh
ADDR=:8080
DB_PATH=yube.db
```
