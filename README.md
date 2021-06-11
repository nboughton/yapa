# yapa: Yet Another Podcast App

yapa is my super basic podcast aggregator that does exactly what I want. The use case for this app is to listen to podcasts in episode order and resume exactly where you left off.

Yapa depends on MPV to actually play episodes.

## What do?

yapa automatically sorts episodes from oldest to newest and, by default, plays the feed in date order. It then marks each episode played at the end of the file and next time you play the feed it picks up at the oldest unplayed episode. If you hit ctrl+c during an episode it notes when you left off and will resume at that point the next time that episode is played.

## Install

At some point I might provide an AUR package for Arch Linux. At the moment you can install it with Go with 

```
go install github.com/nboughton/yapa
```

Then create a config file at ~/.config/yapa/config.json

```
{
	"store": "~/.config/yapa/store.json"
}
```

You can check the usage with 

```
$ yapa -h
A basic podcast aggregator and player for listening to podcasts in episode order

Usage:
  yapa [command]

Available Commands:
  add         Load a new RSS feed to the store
  details     Print details of a feed or episode
  help        Help about any command
  list        List loaded feeds
  mark        Mark toggles episodes played or unplayed
  play        Play a feed, episode or range/set of episodes
  update      Update the store

Flags:
      --config string   config file (default is $HOME/.config/yapa/config.json)
  -h, --help            help for yapa

Use "yapa [command] --help" for more information about a command.
```