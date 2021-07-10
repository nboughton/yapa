/*
Copyright © 2021 Nick Boughton <nicholasboughton@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/nboughton/yapa/pod"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a feed or playlist from the store",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			feed, _     = cmd.Flags().GetInt("feed")
			playlist, _ = cmd.Flags().GetString("playlist")
		)

		if feed < 0 {
			fmt.Println("Please specify a feed.")
			return
		}
		if err := validFeed(feed); err != nil {
			fmt.Println(err)
			return
		}

		// Delete a playlist
		if playlist != "" {
			if _, ok := store.Feeds[feed].Playlists[playlist]; !ok {
				fmt.Println("Playlist not found.")
				return
			} else {
				fmt.Printf("Delete playist '%s', ", playlist)
				if confirm() {
					delete(store.Feeds[feed].Playlists, playlist)
					pod.WriteStore(store)
					fmt.Printf("Playlist '%s' deleted.\n", playlist)
				}
				return
			}
		}

		title := store.Feeds[feed].Title
		fmt.Printf("Delete feed '%s', ", title)
		if confirm() {
			store.Feeds = append(store.Feeds[:feed], store.Feeds[feed+1:]...)
			fmt.Printf("Feed '%s' deleted.\n", title)
			pod.WriteStore(store)
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().IntP("feed", "f", -1, "Feed to delete.")
	deleteCmd.Flags().StringP("playlist", "l", "", "Playlist to delete.")
}

func confirm() bool {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("are you sure? [y/N]: ")
	text, _ := reader.ReadString('\n')
	return text == "y\n"
}
