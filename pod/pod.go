package pod

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"github.com/mmcdole/gofeed"
)

const (
	ErrorInvalidPath       = "cannot read or create store at given path"
	ErrorStoreDoesNotExist = "no existing store found"
)

// Store is the feed collection, we load it on open and write it on changes
type Store struct {
	Path  string
	Feeds Feeds `json:"feeds"`
}

// ReadStore reads in the json data
func ReadStore(path string) (*Store, error) {
	store := &Store{
		// Replace ~/ with home dir
		Path:  strings.Replace(path, "~/", os.Getenv("HOME")+"/", 1),
		Feeds: Feeds{},
	}

	// Validate dirpath
	dir := filepath.Dir(store.Path)
	if err := os.MkdirAll(dir, 0770); err != nil {
		return nil, fmt.Errorf(ErrorInvalidPath)
	}

	// Attempt to open a file there
	f, err := os.Open(store.Path)
	if err != nil {
		// Return an empty store if none exists
		return store, fmt.Errorf(ErrorStoreDoesNotExist)
	}
	defer f.Close()

	// Read store into a Store struct
	err = json.NewDecoder(f).Decode(&store.Feeds)
	return store, err
}

// WriteStore writes to the json store
func WriteStore(store *Store) error {
	f, err := os.Create(store.Path)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(&store.Feeds)
}

// Exists checks for an existing version of a feed in the store
func (store *Store) Exists(name string) bool {
	for _, f := range store.Feeds {
		if f.Title == name {
			return true
		}
	}

	return false
}

// Update the store
func (store *Store) Update() error {
	for _, f := range store.Feeds {
		fmt.Printf("Checking %s\n", f.Title)
		if err := f.Update(); err != nil {
			log.Printf("\tUpdate error: %s\n", err)
		}
	}

	sort.Sort(store.Feeds)
	return nil
}

// Feed data
type Feed struct {
	Title    string    `json:"title"`
	URL      string    `json:"url"`
	RSS      string    `json:"rss"`
	Updated  time.Time `json:"updated"`
	Episodes Episodes  `json:"episodes"`
}

// Played episodes
func (f *Feed) Played() int {
	n := 0
	for _, ep := range f.Episodes {
		if ep.Played {
			n++
		}
	}
	return n
}

// String implements the Stringer interface
func (f *Feed) String() string {
	return fmt.Sprintf("Title:\t%s\nURL:\t%s\nRSS:\t%s\nUpdated:\t%s\nEpisodes:\t%d/%d\n",
		f.Title, f.URL, f.RSS, f.Updated.Format("2006-01-02"), len(f.Episodes), f.Played())
}

// Feed list sortable by most reent update
type Feeds []Feed

// Implement sort interface by last update for Feeds
func (f Feeds) Len() int           { return len(f) }
func (f Feeds) Less(i, j int) bool { return f[i].Updated.After(f[j].Updated) }
func (f Feeds) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

// Update the feed
func (f *Feed) Update() error {
	latest, err := FromRSS(f.RSS)
	if err != nil {
		return err
	}

	// Check if the latest publish date is different.
	// Since we sort oldest to newest by default new episodes should only appear at the end
	if latest.Updated.After(f.Updated) || len(latest.Episodes) != len(f.Episodes) {
		for i, ep := range latest.Episodes {
			if i < len(f.Episodes) && f.Episodes[i].Published != ep.Published {
				return fmt.Errorf("data mismatch:\nold: %+v\nnew: %+v", f.Episodes[i], ep)
			} else if i >= len(f.Episodes) {
				fmt.Printf("\tAdding episode: %s\n", ep.Title)
				f.Episodes = append(f.Episodes, ep)
			}
		}
	}

	return nil
}

// Episode data
type Episode struct {
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Mp3       string    `json:"mp3"`
	Published time.Time `json:"published"`
	Played    bool      `json:"played"`
	Elapsed   int       `json:"elapsed"`
	// Desc      string    `json:"desc"`
}

// String implements the Stringer interface
func (e *Episode) String() string {
	return fmt.Sprintf("Title:\t%s\nURL:\t%s\nMP3:\t%s\nUpdated:\t%s\nPlayed:\t%v\n",
		e.Title, e.URL, e.Mp3, e.Published.Format("2006-01-02"), e.Played)
}

// Play an episode
/* oto segfaults on ctx.Close()
func (e *Episode) Play() error {
	// Get MP3 url
	req, err := http.Get(e.Mp3)
	if err != nil {
		return err
	}
	defer req.Body.Close()

	dec, err := mp3.NewDecoder(req.Body)
	if err != nil {
		return err
	}

	ctx, err := oto.NewContext(dec.SampleRate(), 2, 2, 4096)
	if err != nil {
		return err
	}
	defer ctx.Close()

	plr := ctx.NewPlayer()
	defer plr.Close()

	fmt.Printf("Episode: %s\n", e.Title)
	if _, err := io.Copy(plr, dec); err != nil {
		return err
	}

	// Dont forget to write the store after each succesful play
	e.Played = true
	return nil
}
*/

// Play an episode with mpv
func (e *Episode) PlayMpv() error {
	fmt.Printf("Episode: %s\n", e.Title)

	var cmd *exec.Cmd
	if e.Elapsed > 0 {
		cmd = exec.Command("mpv", "--no-video", e.Mp3, fmt.Sprintf("--start=%d", e.Elapsed))
	} else {
		cmd = exec.Command("mpv", "--no-video", e.Mp3)
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Record time elapsed
	tick := time.NewTicker(time.Second)
	done := make(chan bool, 1)
	defer close(done)

	go func() {
		for {
			select {
			case <-tick.C:
				e.Elapsed++
			case <-done:
				return
			}
		}
	}()

	// Catch kill signal and record elapsed time before closing
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT}...)

		// Block until signal
		s := <-ch

		// Stop tick
		tick.Stop()
		done <- true

		// Send signal to process, this will get returned as err from cmd.Wait below
		cmd.Process.Signal(s)
	}()

	if err := cmd.Wait(); err != nil {
		return err
	}

	// Tidy up if the epsidoe is played completely
	tick.Stop()
	done <- true

	e.Played = true
	e.Elapsed = 0

	return nil
}

// Episodes is its own type in order to implement a sort interface
type Episodes []*Episode

// Implement sort interface by publish date for Espisodes
func (e Episodes) Len() int           { return len(e) }
func (e Episodes) Less(i, j int) bool { return e[i].Published.Before(e[j].Published) }
func (e Episodes) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }

// FromRSS creates a new Feed obj by parsing data from an rss url
func FromRSS(url string) (Feed, error) {
	var fd Feed

	// Parse feed
	fp := gofeed.NewParser()
	f, err := fp.ParseURL(url)
	if err != nil {
		return fd, err
	}

	// Load key data to Feed obj
	fd = Feed{
		Title:   f.Title,
		URL:     f.Link,
		RSS:     url, // Use the passed URL as that will contain auth info if there is any
		Updated: *f.PublishedParsed,
	}

	for _, item := range f.Items {
		fd.Episodes = append(fd.Episodes, &Episode{
			Title:     item.Title,
			URL:       item.Link,
			Mp3:       item.Enclosures[0].URL,
			Published: *item.PublishedParsed,
			// Desc:      item.Description,
		})
	}

	// Default sort by oldest first
	sort.Sort(fd.Episodes)

	return fd, nil
}
