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

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List feeds in store",
	//Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		feed, _ := cmd.Flags().GetInt("feed")
		if feed < 0 {
			fmt.Fprint(tw, "ID\tName\tEps\tPlayed\tLast Updated\n")
			for i, feed := range store.Feeds {
				fmt.Fprintf(tw, "%d\t%s\t%d\t%d\t%s\n", i, feed.Title, len(feed.Episodes), feed.Played(), feed.Updated.Format(dateFmt))
			}
			tw.Flush()
			return
		}

		fmt.Fprint(tw, "ID\tName\tPlayed\tPub Date\n")
		for i, ep := range store.Feeds[feed].Episodes {
			if ep.Played {
				fmt.Fprintf(tw, "%d\t%s\tYes\t%s\n", i, ep.Title, ep.Published.Format(dateFmt))
			} else {
				fmt.Fprintf(tw, "%d\t%s\tNo\t%s\n", i, ep.Title, ep.Published.Format(dateFmt))
			}
		}
		tw.Flush()
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().IntP("feed", "f", -1, "List episodes for feed")
}
