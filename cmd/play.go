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
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
	"github.com/nboughton/yapa/pod"
	"github.com/spf13/cobra"
)

// playCmd represents the play command
var playCmd = &cobra.Command{
	Use:   "play",
	Short: "Play a feed, episode or range/set of episodes",
	//Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		var (
			feed, _     = cmd.Flags().GetInt("feed")
			episodes, _ = cmd.Flags().GetString("episodes")
			speed, _    = cmd.Flags().GetFloat32("speed")
			feedTitle   = store.Feeds[feed].Title
			playlist, _ = cmd.Flags().GetString("playlist")
		)

		if playlist != "" {
			list, ok := store.Feeds[feed].Playlists[playlist]
			if ok {
				for _, id := range list {
					play(store.Feeds[feed].Episodes[id], feedTitle, speed, true)
				}
				return
			} else {
				log.Fatalf("invalid playlist: [%s]", playlist)
			}
		}

		if episodes == "" {
			for _, ep := range store.Feeds[feed].Episodes {
				play(ep, feedTitle, speed, true)
			}
			return
		}

		switch {
		case epSingle.MatchString(episodes):
			id, _ := strconv.Atoi(episodes)

			// Single eps will always play regardless of mark
			if id < len(store.Feeds[feed].Episodes) {
				play(store.Feeds[feed].Episodes[id], feedTitle, speed, false)
			} else {
				log.Fatalf("invalid episode id [%d]", id)
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
				play(ep, feedTitle, speed, true)
			}

		case epSet.MatchString(episodes):
			set := strings.Split(episodes, ",")

			for _, i := range set {
				id, _ := strconv.Atoi(i)
				if id < len(store.Feeds[feed].Episodes) {
					play(store.Feeds[feed].Episodes[id], feedTitle, speed, true)
				}
			}

		default:
			log.Fatalf("Bad criteria: %s", episodes)
		}
	},
}

func init() {
	rootCmd.AddCommand(playCmd)

	playCmd.Flags().IntP("feed", "f", 0, "Play feed, by default episodes marked played are ignored")
	playCmd.Flags().StringP("episodes", "e", "", "Episode or set of episodes to play. Use a single id, a hyphenated pair of ids (0-4), or a comma separated set of ids (0,5,3). Sets cannot have spaces.")
	playCmd.Flags().StringP("playlist", "p", "", "Play a saved playlist. Use the details subcommand to see if a feed has any saved lists.")
	playCmd.Flags().Float32P("speed", "s", 1.0, "Play speed. Accepts values from 0.01 to 100")
}

func play(ep *pod.Episode, feedTitle string, playSpeed float32, skipPlayed bool) {
	if skipPlayed && ep.Played {
		return
	}

	if showNotify {
		go sendNotify(feedTitle, ep.Title)
	}

	args := []string{
		"--no-video",
		ep.Mp3,
		fmt.Sprintf("--speed=%.2f", playSpeed),
	}
	if ep.Elapsed > 0 {
		args = append(args, fmt.Sprintf("--start=%d", ep.Elapsed))

		for _, i := range []int{3, 2, 1} {
			clear()
			fmt.Printf("Feed: %s\nPlaying: %s\n", feedTitle, ep.Title)
			fmt.Printf("-> Resuming at %s in %d\n", pod.ParseElapsed(ep.Elapsed), i)
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
				fmt.Printf("Feed: %s\nPlaying: %s\nElapsed: %s\n", feedTitle, ep.Title, pod.ParseElapsed(ep.Elapsed))
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
		os.Exit(1)
	}

	// Tidy up if the epsiode is played completely
	done <- true

	ep.Played = true
	ep.Elapsed = 0
	pod.WriteStore(store)
}

func clear() error {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	return cmd.Run()
}

func sendNotify(feed, episode string) error {
	conn, err := dbus.SessionBusPrivate()
	if err != nil {
		return err
	}
	defer conn.Close()

	if err = conn.Auth(nil); err != nil {
		return err
	}

	if err = conn.Hello(); err != nil {
		return err
	}

	// Send notification
	_, err = notify.SendNotification(conn, notify.Notification{
		AppName:       "yapa",
		ReplacesID:    uint32(0),
		Summary:       "Podcast",
		Body:          fmt.Sprintf("%s\n%s\n", feed, episode),
		Hints:         map[string]dbus.Variant{},
		ExpireTimeout: 10000,
	})

	return err
}
