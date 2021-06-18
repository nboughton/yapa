# yapa: Yet Another Podcast App

yapa is my super basic podcast aggregator that does exactly what I want. The use case for this app is to listen to podcasts in episode order and resume exactly where you left off.

Yapa depends on MPV to actually play episodes.

## What do?

yapa automatically sorts episodes from oldest to newest and, by default, plays the feed in date order. It then marks each episode played at the end of the file and next time you play the feed it picks up at the oldest unplayed episode. If you hit ctrl+c during an episode it notes when you left off and will resume at that point the next time that episode is played.

Yapa is *very* basic. It stores feed data as a JSON file that is read when the yapa command is invoked and written on any change. Don't try to update the store while yapa is already playing as the changes will be overwritten when the store is updated after each episode.

## Install

For Arch Linux:
```
yay -S yapa
```

Or you can build it from source locally:

```
go install github.com/nboughton/yapa
```

Then create a config file at ~/.config/yapa/config.json

```
{
	"store": "~/.config/yapa/store.json",
  "notify": true
}
```
You can disable desktop notifications by setting notify to false.

Now add a feed:

```
yapa add <RSS feed url>
```

For help see: 

```
$ yapa -h
A basic podcast aggregator and player for listening to podcasts in episode order

Usage:
  yapa [command]

Available Commands:
  add         Load a new RSS feed to the store
  details     Print details of a feed, episode or range/set of episodes
  help        Help about any command
  list        List feeds in store
  mark        Mark an episode or range/set of episodes played/unplayed
  play        Play a feed, episode or range/set of episodes
  update      Update the store

Flags:
      --config string   config file (default is $HOME/.config/yapa/config.json)
  -h, --help            help for yapa

Use "yapa [command] --help" for more information about a command.
```

## Example

```
❯ yapa list

ID  Name                           Eps  Played  Last Updated
0   RQ Early Access Patron Feed    690  0       2021-06-09 23:18
1   Rusty Quill Gaming Podcast     294  7       2021-06-09 15:00
2   This Paranormal Life           218  218     2021-06-08 22:38
3   Dark Air with Terry Carnation  12   5       2021-06-08 19:25
4   D&D is For Nerds               350  0       2021-06-05 14:00
5   Stellar Firma                  119  9       2021-06-04 15:00
6   The Magnus Archives            261  0       2021-06-03 15:00
7   Hearty Dice Friends            204  0       2021-05-14 15:39
8   Power Word Roll                70   0       2021-01-28 11:00

~
❯ yapa play -f5

Feed: Stellar Firma
Playing: Episode 8 - Pillows and Cults
-> Resuming at 12m 33s
```

## Playlists

You can filter episodes with the list command like so:

```
yapa list -f0 -r '^regex$'
```

The filter option (-r/--filter) takes a string that should be an RE2 compatible regular expression. Once your happy with the filtered list you can add the -s/--save flag like so:

```
yapa list -f0 -r '^regex$' -s 'Playlist Name'
```

You can then play that list later with:

```
yapa play -f0 -p'Playlist Name'
```

You can check for saved playlists with the details command:

```
yapa details -f0
```