# yapa: Yet Another Podcast App

yapa is my super basic podcast aggregator that does exactly what I want. It uses mpv to
play files. Please don't submit feature requests. I wrote this just for me and I'm not interested in anyone elses requirements.

## What do?

yapa has some subcommands that allow you to add, list, and play feeds. It automatically
sorts episodes from oldest to newest and plays the feed in date order. It then marks each
episode played at the end of the file and next time you play the feed it picks up at the oldest unplayed episode.

## Why are you still reading this?

Seriously this is a really bad app. You can install it with Go with 

```
go install github.com/nboughton/yapa
```