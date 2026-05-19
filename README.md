# TubeFeed

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
DB_PATH=youtube-feed-reader.db
```



# Updates

<!-- - regardless of image aspect ratio, all images should be the same size and zoom cropped if needed
- images on channels and main page are slightly different sizes, they should be the same
- Switch to the ORM bun
- Start grouping routes in server.go
- Make settings functional
- Update the html and css to use html5 elements where reasonable instead of divs.  Think like article, section, header, footer.  Also focus on reusable templates and break up the templates into simpler smaller ones as needed. -->
<!-- - Make feed page look better, fix opml import, want it too look modern but remember opml import is really a one time thing.
- Search loses the query when paginating, also fix pagination buttons on search and main page to look better
- The nav should have the same selected yellow background as the channels when selected
- on feeds the import opml and add feed should be next to each other and moved to the bottom under feeds.  Feeds should be clickable and link to the channel.  Feed across the app should be renamed to Channels.  Overall this shouldn't be a generic opml app, but rather a youtube specific app, terminolgy should reflect that globally.
- Channels page should fill out the rest of the data in the table.
- Channels view should paginate.
- Channels title should be the name of the channel instead of Latest videos
- db.go doesn't seem to be fully using bun, with may raw queries.  -->

<!-- - each channel should show an unsubscribe button in channels -->
<!-- - Import subscriptions should allow files to be dropped on it and look nicer, the current file picker is ugly clicking on the drop area should open the file picker.
- We're saving the same url prefixes for videos and thumbnails, seems wasteful we really only need the unique part and can construct the urls in code.
- Remove the google fon't reference add local fonts if needed -->
- checkmark should not be green, i've already updated this.
- Last updated time should update when refresh button is clicked
- buttons in topbar should have hover state like next prev buttons
- The yellow highlight color is ugly and looks old fashioned, we should use something that reflects the logo
- should the moon and sun icon be flipped?
- when refresh is clicked the icon should spin twice.
- last updated should update immediatley when refresh is clicked.
- replace the ☰ icon with a font icon
