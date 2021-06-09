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

If, for some ungodly reason, you actually want to use it then you'll want to create a config file at ~/.config/yapa/config.json with the following template:

```
{
	"db": "~/.config/yapa/db.json"
}
```

You can check the usage with 

```
$yapa -h
Prototype pocasting app with next ep autoplay

Usage:
  yapa [command]

Available Commands:
  add         Load a new RSS feed to the database
  help        Help about any command
  list        List loaded feeds
  play        Play a feed or episode
  update      Update the store

Flags:
      --config string   config file (default is $HOME/.config/yapa/config.json)
  -h, --help            help for yapa

Use "yapa [command] --help" for more information about a command.
```