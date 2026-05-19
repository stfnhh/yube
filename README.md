# Yubè

Yubè is a tiny self-hosted channel reader for people who want a quieter way to keep up with videos from the channels they follow. It pulls channel updates into a local SQLite database, shows new videos in a clean web UI, tracks what you have watched, and exposes a unified RSS feed for other apps.

It is intentionally small: no accounts, no recommendation engine, no social layer, and no attempt to recreate the full video platform. It is a personal subscriptions inbox that you can run on your own machine or server.

## What It Does

- Follow channels by pasting a channel URL, handle URL, or Atom URL.
- Import subscriptions from an OPML/XML file.
- Show a responsive card grid of recent videos.
- Track watched videos and reduce the channel unread count automatically.
- Browse a single channel from the sidebar or Channels page.
- Search across channels and videos.
- Refresh channels automatically in the background.
- Refresh manually from the top bar.
- Keep local channel icons and video thumbnails available through the app.
- Configure refresh and cleanup behavior from Settings.
- Expose a unified RSS feed for apps like Miniflux, FreshRSS, NetNewsWire, Reeder, or any other RSS reader.
- Store everything in a single SQLite database that is easy to back up.

## Screens And Concepts

### Watch

The main page shows the latest videos across all channels. Clicking a video opens the original video in a new tab and marks it as watched in Yubè.

Watched state is local to Yubè. It is used to update the checkmark on video cards and the unread counts in the sidebar.

### Channels

The Channels page lists every subscribed channel, including its unwatched video count. From here you can:

- Open a channel-specific view.
- Add a new channel.
- Import subscriptions from OPML/XML.
- Unsubscribe from a channel after confirmation.

### Search

Search has its own route at `/search?q=...`. It searches both videos and channels. Channel results appear alongside video results so you can jump directly to a channel.

### Settings

Settings lets you tune the app without editing config files:

- Refresh interval.
- Video retention.
- Maximum videos kept per channel.

The app applies refresh settings without needing a restart.

## RSS For Other Apps

Yubè exposes one unified RSS feed containing the latest videos across all channels:

```text
/rss.xml
```

The RSS feed includes:

- RSS 2.0 output.
- Atom self-link metadata.
- Media RSS thumbnail metadata.
- CDATA descriptions with thumbnail, channel name, and title.
- Stable item GUIDs based on video IDs.
- Publish dates from the source channel.

Example RSS URL:

```text
http://your-server:8080/rss.xml
```

If you run Yubè behind a reverse proxy, make sure the proxy forwards `X-Forwarded-Proto` and `X-Forwarded-Host`. Yubè uses those headers when building absolute RSS thumbnail URLs.

## Running With Docker

Build the image locally:

```sh
docker build -t yube .
```

Run it with persistent storage:

```sh
docker run -d \
  --name yube \
  -p 8080:8080 \
  -v yube-data:/data \
  --restart unless-stopped \
  yube
```

Open:

```text
http://localhost:8080
```

The Docker image stores the database at:

```text
/data/yube.db
```

Keep `/data` persistent. If that volume is removed, your subscriptions, watched state, settings, icons, and stored video metadata are removed with it.

## Docker Compose

A minimal Compose setup:

```yaml
services:
  yube:
    build: .
    container_name: yube
    ports:
      - "8080:8080"
    volumes:
      - yube-data:/data
    restart: unless-stopped

volumes:
  yube-data:
```

Then start it:

```sh
docker compose up -d
```

## Configuration

Yubè only needs a couple of environment variables for hosting:

| Variable | Default | Description |
| --- | --- | --- |
| `ADDR` | `:8080` | Address and port the web server listens on. |
| `DB_PATH` | `yube.db` locally, `/data/yube.db` in Docker | SQLite database path. |

Example:

```sh
docker run -d \
  --name yube \
  -p 9090:9090 \
  -e ADDR=:9090 \
  -v yube-data:/data \
  yube
```

Most day-to-day behavior is configured in the Settings page rather than through environment variables.

## Adding Channels

Yubè accepts channel-style URLs and direct Atom URLs. Common inputs include:

```text
https://www.youtube.com/@somechannel
https://www.youtube.com/channel/CHANNEL_ID
https://www.youtube.com/feeds/videos.xml?channel_id=CHANNEL_ID
```

When possible, Yubè resolves the channel ID, stores the channel name, fetches the channel icon, and begins importing videos.

## Importing Subscriptions

On the Channels page, use Import subscriptions to upload an OPML or XML file. You can click the drop area or drag a file onto it.

This is useful if you exported subscriptions from another reader or from a subscriptions manager that can produce OPML.

## Backups

The important file is the SQLite database:

```text
yube.db
```

In Docker, it lives in the mounted `/data` volume. Back up the database or the whole Docker volume.

For safest backups, stop the container first:

```sh
docker stop yube
```

Then copy or snapshot the volume/database, and start it again:

```sh
docker start yube
```

SQLite may also create `-wal` and `-shm` files while the app is running. If you back up a live database, include those files too.

## Reverse Proxy Notes

Yubè does not currently include built-in authentication. If you expose it outside your home network, put it behind something that handles access control, such as:

- A VPN.
- Tailscale or a similar private network.
- A reverse proxy with authentication.
- Your existing home-lab single sign-on setup.

For RSS thumbnail URLs to be generated correctly behind a proxy, forward these headers:

```text
X-Forwarded-Proto
X-Forwarded-Host
```

## Limitations

Yubè is deliberately narrow in scope:

- It is not a multi-user app.
- It does not sync watch history back to the source platform.
- It does not download or mirror videos.
- It does not include comments, recommendations, playlists, or notifications.
- It depends on public channel metadata continuing to work.

That narrowness is the point: it is meant to be a calm subscriptions inbox, not another infinite feed.

## A Note On How This Was Built

This project was almost exclusively vibe coded.

That is not meant as a badge of engineering purity. It started as a personal tool to solve a very specific annoyance: I wanted a simple, self-hosted way to follow channels, see what was new, and keep that data available to RSS readers without turning the experience into a full social/video platform.

Because it was built primarily to solve my own problem, some choices may be more practical than polished, and the app may not cover every workflow someone else expects. Still, I suspect the same shape of problem exists for other self-hosters: a small local subscriptions inbox, readable through the web or RSS, with the data under your control.

If that is useful to you too, excellent. That is the spirit of the project.

## Local Non-Docker Run

Docker is the recommended path for self-hosting, but you can also run it directly if Go is installed:

```sh
go run ./cmd/server
```

Then open:

```text
http://localhost:8080
```

You can choose a custom database path or port:

```sh
DB_PATH=/path/to/yube.db ADDR=:8080 go run ./cmd/server
```
