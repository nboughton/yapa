# yapa: Yet Another Podcast App

yapa is my super basic podcast aggregator that does exactly what I want. The use case for this app is to listen to podcasts in episode order and resume exactly where you left off.

Yapa depends on MPV to actually play episodes.

## What do?

yapa automatically sorts episodes from oldest to newest and, by default, plays the feed in date order. It then marks each episode played at the end of the file and next time you play the feed it picks up at the oldest unplayed episode. If you hit ctrl+c during an episode it notes when you left off and will resume at that point the next time that episode is played.

Yapa is *very* basic. It stores feed data as a JSON file that is read when the yapa command is invoked and written on any change. Don't try to update the store while yapa is already playing as the changes will be overwritten when the store is updated after each episode.

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
Episode: Episode 8 - Pillows and Cults
	Resuming at 12m 33s
```

## Playlists?

Well... Sort of? I make playlists like so:

```
$ yapa list -f6 | grep term | awk '{print $1}' | paste -sd ','

248,249,250,251,252,253,254,255,256,257,258,259,260,261,262,263,264,265,266,267,268,269,270,271,272,273,279,280,281,282,283,284,285,286,287,288,289,290,291,292,293,294,295,296,297,298,299,300,301,302,303,315,316,317,318,319,320,321,322,323,324,325,326,327,328,329,330,331,332,333,334,335,336,337,338,339,340
```

Which gives you a comma separated list of episodes that match your term. Leave out the awk and paste pipes to see the episode list so you can refine your grep. You can then play just those episodes with:

```
yapa play -f6 -e248,249,250,251,252,253,254,255,256,257,258,259,260,261,262,263,264,265,266,267,268,269,270,271,272,273,279,280,281,282,283,284,285,286,287,288,289,290,291,292,293,294,295,296,297,298,299,300,301,302,303,315,316,317,318,319,320,321,322,323,324,325,326,327,328,329,330,331,332,333,334,335,336,337,338,339,340
```

Played episodes will be ignored by default and yapa should just pick up wherever you left off.