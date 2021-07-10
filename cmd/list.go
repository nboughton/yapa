// Copyright © 2021 Nick Boughton <nicholasboughton@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package cmd

import (
	"fmt"

	"github.com/nboughton/yapa/pod"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List feeds/episodes in store",
	Long:  `List output can be marked as played/unplayed, and saved as playlists. The --filter, --episodes, and --playlist flags are mutually exclusive.`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			feed, _         = cmd.Flags().GetInt("feed")
			episodes, _     = cmd.Flags().GetString("episodes")
			filter, _       = cmd.Flags().GetString("filter")
			save, _         = cmd.Flags().GetString("save")
			list, _         = cmd.Flags().GetString("playlist")
			add, _          = cmd.Flags().GetString("add-to-playlist")
			summary, _      = cmd.Flags().GetBool("summary")
			details, _      = cmd.Flags().GetBool("details")
			markPlayed, _   = cmd.Flags().GetBool("mark-played")
			markUnplayed, _ = cmd.Flags().GetBool("mark-unplayed")
			playlist        []int
		)

		// Various checks to reject conflicting flags
		if markPlayed && markUnplayed {
			fmt.Println("Please only select mark-played OR mark-unplayed")
			return
		}

		if save != "" && add != "" {
			fmt.Println("You can save a list or append to an existing one. You can't do both.")
			return
		}

		// Check filters
		excFlagCount := 0
		if filter != ".*" {
			excFlagCount++
		}
		if episodes != "" {
			excFlagCount++
		}
		if list != "" {
			excFlagCount++
		}
		if excFlagCount > 1 {
			fmt.Println("--filter (-r), --episodes (-e), and --playlist (-l) are mutually exclusive.")
			return
		}

		// No feed specified, print basic summary of all feeds
		if feed < 0 {
			if !details {
				fmt.Fprint(tw, "ID\tName\tEps\tPlayed\tLast Updated\n")
			}

			for i, feed := range store.Feeds {
				if details {
					fmt.Fprint(tw, feed.String())
				} else {
					fmt.Fprintf(tw, "%d\t%s\t%d\t%d\t%s\n", i, feed.Title, len(feed.Episodes), feed.Played(), feed.Updated.Format(dateFmt))
				}
			}

			tw.Flush()
			return
		}

		// Print summary of selected feed
		if summary {
			fmt.Fprint(tw, store.Feeds[feed].String())
			tw.Flush()
			return
		}

		// List episodes from selected feed, apply mark/print full details if flagged
		if !details {
			fmt.Fprint(tw, "ID\tName\tPlayed\tPub Date\n")
		}

		// Get episode list to process
		var eps pod.Episodes
		switch {
		case filter != ".*":
			eps = store.Feeds[feed].Filter(filter)
		case episodes != "":
			eps = store.Feeds[feed].Set(episodes)
		case list != "":
			eps = store.Feeds[feed].Playlist(list)
		default:
			eps = store.Feeds[feed].Episodes
		}

		// Iterate and process episodes
		for _, ep := range eps {
			if markPlayed {
				ep.Played = true
			}
			if markUnplayed {
				ep.Played = false
			}
			if details {
				fmt.Fprintln(tw, ep)
			} else {
				fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", ep.ID, ep.Title, played(ep.Played), ep.Published.Format(dateFmt))
			}
			if save != "" || add != "" {
				playlist = append(playlist, ep.ID)
			}
		}

		// Write changes to store
		if markPlayed || markUnplayed {
			pod.WriteStore(store)
		}

		// Print list text
		tw.Flush()

		// Save the playlist
		if save != "" {
			// Check the map exists
			if store.Feeds[feed].Playlists == nil {
				store.Feeds[feed].Playlists = make(map[string][]int)
			}
			store.Feeds[feed].Playlists[save] = playlist
			pod.WriteStore(store)

			fmt.Printf("Playlist saved as '%s'\n", save)
		}

		// Append to an existing playlist
		if add != "" {
			if _, ok := store.Feeds[feed].Playlists[add]; ok {
				store.Feeds[feed].Playlists[add] = append(store.Feeds[feed].Playlists[add], playlist...)
				pod.WriteStore(store)

				fmt.Printf("Episodes appended to '%s' playlist\n", add)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().IntP("feed", "f", -1, "List episodes for feed")
	listCmd.Flags().StringP("filter", "r", ".*", "Filter episodes with a RE2 compatible regular expression.")
	listCmd.Flags().StringP("episodes", "e", "", "Filter episodes as a range (0-10) or a comma separated set (3,5,6). No spaces.")
	listCmd.Flags().StringP("playlist", "l", "", "Print playlist.")
	listCmd.Flags().StringP("save", "s", "", "Save results as playlist.")
	listCmd.Flags().StringP("add-to-playlist", "a", "", "Append episodes to an existing playlist.")
	listCmd.Flags().BoolP("summary", "m", false, "Only print summary for selected feed.")
	listCmd.Flags().BoolP("details", "d", false, "Print full details of selected feed/episode.")
	listCmd.Flags().BoolP("mark-played", "p", false, "Mark the listed episodes as played.")
	listCmd.Flags().BoolP("mark-unplayed", "u", false, "Mark the listed episodes as unplayed.")
}

func played(p bool) string {
	if p {
		return "Yes"
	}

	return "No"
}
