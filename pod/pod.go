package pod

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
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
		if err := f.Update(); err != nil {
			return err
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

// Play the feed
// This isn't great because we can't write the store after each play.
/*
func (f *Feed) Play() error {
	fmt.Printf("Feed: %s\n", f.Title)
	for _, ep := range f.Episodes {
		if !ep.Played {
			if err := ep.PlayMpv(); err != nil {
				return err
			}
		}
	}

	return nil
}
*/

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
	// Since we sort oldest to newset by default new episodes should only appear at the end
	if latest.Updated.After(f.Updated) || len(latest.Episodes) != len(f.Episodes) {
		for i, ep := range latest.Episodes {
			if i < len(f.Episodes) && f.Episodes[i].Published != ep.Published {
				return fmt.Errorf("data mismatch:\nold: %+v\nnew: %+v", f.Episodes[i], ep)
			} else if i >= len(f.Episodes) {
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
	// Desc      string    `json:"desc"`
}

// Play an episode
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

func (e *Episode) PlayMpv() error {
	fmt.Printf("Episode: %s\n", e.Title)
	if _, err := exec.Command("mpv", "--no-video", e.Mp3).CombinedOutput(); err != nil {
		return err
	}

	e.Played = true
	return nil
}

// Episodes is its own type in order to impement a sort interface
type Episodes []*Episode

// Implement sort interface by pulish date for Espisodes
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
		RSS:     f.FeedLink,
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
