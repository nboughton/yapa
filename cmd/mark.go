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
	Short: "Mark episodes played or unplayed, played episodes will be marked unplayed and vice versa",
	//Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		f, _ := cmd.Flags().GetInt("feed")
		e, _ := cmd.Flags().GetString("episodes")

		if f == -1 {
			log.Fatal("No feed selected")
		}

		switch {
		case epSingle.MatchString(e):
			n, _ := strconv.Atoi(e)
			store.Feeds[f].Episodes[n].Played = !store.Feeds[f].Episodes[n].Played
			pod.WriteStore(store)

		case epRange.MatchString(e):
			r := strings.Split(e, "-")
			start, _ := strconv.Atoi(r[0])
			end, _ := strconv.Atoi(r[1])
			if end+1 > len(store.Feeds[f].Episodes) {
				end = len(store.Feeds[f].Episodes) - 1
			}
			for _, ep := range store.Feeds[f].Episodes[start : end+1] {
				ep.Played = !ep.Played
			}
			pod.WriteStore(store)

		case epSet.MatchString(e):
			r := strings.Split(e, ",")
			for _, i := range r {
				d, _ := strconv.Atoi(i)
				if d < len(store.Feeds[f].Episodes) {
					store.Feeds[f].Episodes[d].Played = !store.Feeds[f].Episodes[d].Played
				}
			}
			pod.WriteStore(store)

		default:
			log.Fatalf("Bad criteria: %s", e)
		}
	},
}

func init() {
	rootCmd.AddCommand(markCmd)

	markCmd.Flags().IntP("feed", "f", -1, "Feed id to mark")
	markCmd.Flags().StringP("episodes", "e", "", "Episode or set of episodes to mark. Use a single id, a hyphenated pair of ids (0-4), or a comma separated set of ids (0,5,3). Sets should have no spaces.")
}
