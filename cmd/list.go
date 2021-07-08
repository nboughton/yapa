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
	Long:  `List output can be marked as played/unplayed, and saved as playlists`,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			feed, _         = cmd.Flags().GetInt("feed")
			episodes, _     = cmd.Flags().GetString("episodes")
			filter, _       = cmd.Flags().GetString("filter")
			save, _         = cmd.Flags().GetString("save")
			playlists, _    = cmd.Flags().GetBool("playlists")
			details, _      = cmd.Flags().GetBool("details")
			markPlayed, _   = cmd.Flags().GetBool("mark-played")
			markUnplayed, _ = cmd.Flags().GetBool("mark-unplayed")
			playlist        []int
		)

		if markPlayed && markUnplayed {
			fmt.Println("Please only select mark-played OR mark-unplayed")
			return
		}

		if feed < 0 {
			fmt.Fprint(tw, "ID\tName\tEps\tPlayed\tLast Updated\n")
			for i, feed := range store.Feeds {
				if details {
					fmt.Println(feed.String())
				} else {
					fmt.Fprintf(tw, "%d\t%s\t%d\t%d\t%s\n", i, feed.Title, len(feed.Episodes), feed.Played(), feed.Updated.Format(dateFmt))
				}
			}

			if !details {
				tw.Flush()
			}
			return
		}

		if playlists {
			fmt.Fprintln(tw, store.Feeds[feed].String())
			tw.Flush()
			return
		}

		// List episodes from selected feed, apply mark/print full details if flagged
		fmt.Fprint(tw, "ID\tName\tPlayed\tPub Date\n")
		var eps pod.Episodes
		switch {
		case filter != ".*":
			eps = store.Feeds[feed].Filter(filter)
		case episodes != "":
			eps = store.Feeds[feed].Set(episodes)
		default:
			eps = store.Feeds[feed].Episodes
		}

		for _, ep := range eps {
			if markPlayed {
				ep.Played = true
			}
			if markUnplayed {
				ep.Played = false
			}
			if details {
				fmt.Println(ep)
			} else {
				fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", ep.ID, ep.Title, played(ep.Played), ep.Published.Format(dateFmt))
			}
			if save != "" {
				playlist = append(playlist, ep.ID)
			}
		}

		// Write changes to store
		if markPlayed || markUnplayed {
			pod.WriteStore(store)
		}

		// Print list text
		if !details {
			tw.Flush()
		}

		// Save the playlist
		if save != "" {
			// Check the map exists
			if store.Feeds[feed].Playlists == nil {
				store.Feeds[feed].Playlists = make(map[string][]int)
			}
			store.Feeds[feed].Playlists[save] = playlist
			pod.WriteStore(store)

			fmt.Println("Playlist saved")
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().IntP("feed", "f", -1, "List episodes for feed")
	listCmd.Flags().StringP("filter", "r", ".*", "Filter episodes with a regular expression. See the RE2 specification for details. Use single quotes to wrap your expression.")
	listCmd.Flags().StringP("episodes", "e", "", "Filter episodes as a range (0-10) or a comma separated set (3,5,6). This cannot contain spaces and is overridden by the -r flag.")
	listCmd.Flags().StringP("save", "s", "", "Save list as playlist. You must specify a feed with -f to use this. Playlists are saved as part of the feed record in the store.")
	listCmd.Flags().BoolP("playlists", "l", false, "Only print details/playlists for selected feed.")
	listCmd.Flags().BoolP("details", "d", false, "Print full details of selected feed/episode.")
	listCmd.Flags().BoolP("mark-played", "p", false, "Mark the listed episodes as played. Only works in conjunction with the -f flag.")
	listCmd.Flags().BoolP("mark-unplayed", "u", false, "Mark the listed episodes as unplayed. Only works in conjunction with the -f flag.")
}

func played(p bool) string {
	if p {
		return "Yes"
	}

	return "No"
}
