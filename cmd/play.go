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
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/nboughton/yapa/pod"
	"github.com/spf13/cobra"
)

// playCmd represents the play command
var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Play a feed or playlist",
	//Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			feed, _     = cmd.Flags().GetInt("feed")
			speed, _    = cmd.Flags().GetFloat32("speed")
			feedTitle   = store.Feeds[feed].Title
			playlist, _ = cmd.Flags().GetString("playlist")
		)

		if err := validFeed(feed); err != nil {
			fmt.Println(err)
			return
		}

		if playlist != "" {
			list, ok := store.Feeds[feed].Playlists[playlist]
			if ok {
				for _, id := range list {
					play(store.Feeds[feed].Episodes[id], feedTitle, speed, true)
				}
				return
			} else {
				fmt.Printf("invalid playlist: [%s]", playlist)
				return
			}
		}

		for _, ep := range store.Feeds[feed].Episodes {
			play(ep, feedTitle, speed, true)
		}
	},
}

func init() {
	rootCmd.AddCommand(playCmd)

	playCmd.Flags().IntP("feed", "f", 0, "Play feed, by default episodes marked played are ignored")
	playCmd.Flags().StringP("playlist", "l", "", "Play a saved playlist")
	playCmd.Flags().Float32P("speed", "s", 1.0, "Play speed. Accepts values from 0.01 to 100")
}

func play(ep *pod.Episode, feedTitle string, playSpeed float32, skipPlayed bool) {
	if skipPlayed && ep.Played {
		return
	}

	if showNotify {
		go sendNotify(feedTitle, ep.Title)
	}

	// Hide cursor while playing
	tput(hideCursor)
	defer tput(showCursor)

	args := []string{
		"--no-video",
		ep.Mp3,
		fmt.Sprintf("--speed=%.2f", playSpeed),
	}
	if ep.Elapsed > 0 {
		args = append(args, fmt.Sprintf("--start=%d", ep.Elapsed))

		for _, i := range []int{3, 2, 1} {
			clear()
			fmt.Fprintf(tw, "Feed:\t%s\nPlaying:\t%s\n-> Resuming at %s in %d",
				feedTitle, ep.Title, pod.ParseElapsed(ep.Elapsed), i)
			tw.Flush()
			time.Sleep(time.Second * 1)
		}
	}
	cmd := exec.Command("mpv", args...)

	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	// Record time elapsed and catch kill signal
	done := make(chan bool, 1)
	defer close(done)

	go func() {
		tick := time.NewTicker(time.Second)
		defer tick.Stop()

		sig := make(chan os.Signal, 1)
		defer close(sig)

		signal.Notify(sig, []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}...)
		defer signal.Stop(sig)

		for {
			select {
			case <-tick.C:
				ep.Elapsed++
				clear()
				fmt.Fprintf(tw, "Feed:\t%s\nPlaying:\t%s\nElapsed:\t%s",
					feedTitle, ep.Title, pod.ParseElapsed(ep.Elapsed))
				tw.Flush()
			case <-done:
				return
			case s := <-sig:
				cmd.Process.Signal(s)
				return
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		pod.WriteStore(store)
		tput(showCursor)
		os.Exit(0)
	}

	// Tidy up if the epsiode is played completely
	done <- true

	ep.Played = true
	ep.Elapsed = 0
	pod.WriteStore(store)
}
