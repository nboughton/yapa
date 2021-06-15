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
	"log"
	"strconv"
	"strings"

	"github.com/nboughton/yapa/pod"
	"github.com/spf13/cobra"
)

// markCmd represents the mark command
var markCmd = &cobra.Command{
	Use:   "mark",
	Short: "Mark toggles episodes played or unplayed",
	//Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			feed, _     = cmd.Flags().GetInt("feed")
			episodes, _ = cmd.Flags().GetString("episodes")
			played, _   = cmd.Flags().GetBool("played")
			unplayed, _ = cmd.Flags().GetBool("unplayed")
		)

		if feed == -1 {
			log.Fatal("No feed selected")
		}

		switch {
		case epSingle.MatchString(episodes):
			id, _ := strconv.Atoi(episodes)

			if id < len(store.Feeds[feed].Episodes) {
				switch {
				case played:
					store.Feeds[feed].Episodes[id].Played = true
				case unplayed:
					store.Feeds[feed].Episodes[id].Played = false
				default:
					store.Feeds[feed].Episodes[id].Played = !store.Feeds[feed].Episodes[id].Played
				}

				store.Feeds[feed].Episodes[id].Elapsed = 0
				pod.WriteStore(store)
			}

		case epRange.MatchString(episodes):
			var (
				set      = strings.Split(episodes, "-")
				first, _ = strconv.Atoi(set[0])
				last, _  = strconv.Atoi(set[1])
			)

			if last+1 > len(store.Feeds[feed].Episodes) {
				last = len(store.Feeds[feed].Episodes) - 1
			}

			for _, ep := range store.Feeds[feed].Episodes[first : last+1] {
				switch {
				case played:
					ep.Played = true
				case unplayed:
					ep.Played = false
				default:
					ep.Played = !ep.Played
				}

				ep.Elapsed = 0
			}
			pod.WriteStore(store)

		case epSet.MatchString(episodes):
			set := strings.Split(episodes, ",")

			for _, i := range set {
				id, _ := strconv.Atoi(i)
				if id < len(store.Feeds[feed].Episodes) {
					switch {
					case played:
						store.Feeds[feed].Episodes[id].Played = true
					case unplayed:
						store.Feeds[feed].Episodes[id].Played = false
					default:
						store.Feeds[feed].Episodes[id].Played = !store.Feeds[feed].Episodes[id].Played
					}

					store.Feeds[feed].Episodes[id].Elapsed = 0
				}
			}
			pod.WriteStore(store)

		default:
			log.Fatalf("Bad criteria: %s", episodes)
		}
	},
}

func init() {
	rootCmd.AddCommand(markCmd)

	markCmd.Flags().IntP("feed", "f", -1, "Feed id to mark")
	markCmd.Flags().StringP("episodes", "e", "", "Episode or set of episodes to mark. Use a single id, a hyphenated pair of ids (0-4), or a comma separated set of ids (0,5,3). Sets cannot have spaces.")
	markCmd.Flags().BoolP("played", "p", false, "Mark episodes played")
	markCmd.Flags().BoolP("unplayed", "u", false, "Mark episodes unplayed")
}
