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
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/nboughton/yapa/pod"
	"github.com/spf13/cobra"
)

// playCmd represents the play command
var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Play a feed, episode or range/set of episodes",
	//Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		feed, _ := cmd.Flags().GetInt("feed")
		episodes, _ := cmd.Flags().GetString("episodes")
		speed, _ := cmd.Flags().GetFloat32("speed")

		fmt.Printf("Feed: %s\n", store.Feeds[feed].Title)
		if episodes == "" {
			for id, ep := range store.Feeds[feed].Episodes {
				play(ep, feed, id, speed)
			}
			return
		}

		switch {
		case epSingle.MatchString(episodes):
			id, _ := strconv.Atoi(episodes)

			// Single eps will always play regardless of mark
			if id < len(store.Feeds[feed].Episodes) {
				store.Feeds[feed].Episodes[id].Play(feed, id, speed)
			}

		case epRange.MatchString(episodes):
			set := strings.Split(episodes, "-")
			first, _ := strconv.Atoi(set[0])
			last, _ := strconv.Atoi(set[1])

			if last+1 > len(store.Feeds[feed].Episodes) {
				last = len(store.Feeds[feed].Episodes) - 1
			}

			for id, ep := range store.Feeds[feed].Episodes[first : last+1] {
				play(ep, feed, first+id, speed)
			}

		case epSet.MatchString(episodes):
			set := strings.Split(episodes, ",")

			for _, i := range set {
				id, _ := strconv.Atoi(i)
				if id < len(store.Feeds[feed].Episodes) {
					play(store.Feeds[feed].Episodes[id], feed, id, speed)
				}
			}

		default:
			log.Fatalf("Bad criteria: %speed", episodes)
		}
	},
}

func init() {
	rootCmd.AddCommand(playCmd)

	playCmd.Flags().IntP("feed", "f", 0, "Play feed, by default episodes marked played are ignored")
	playCmd.Flags().StringP("episodes", "e", "", "Episode or set of episodes to play. Use a single id, a hyphenated pair of ids (0-4), or a comma separated set of ids (0,5,3). Sets cannot have spaces.")
	playCmd.Flags().Float32P("speed", "s", 1.0, "Play speed. Accepts values from 0.01 to 100")
}

func play(ep *pod.Episode, feed, id int, speed float32) {
	if !ep.Played {
		if err := ep.Play(feed, id, speed); err != nil {
			pod.WriteStore(store)
			os.Exit(1)
		}
		pod.WriteStore(store)
	}
}
