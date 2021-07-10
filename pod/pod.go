package pod

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
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
	for i, f := range store.Feeds {
		fmt.Printf("Updating %s\n", f.Title)
		if err := store.Feeds[i].Update(); err != nil {
			log.Printf("-> Update error: %s\n", err)
		}
	}

	sort.Sort(store.Feeds)
	return nil
}

// Feed data
type Feed struct {
	Title     string           `json:"title"`
	URL       string           `json:"url"`
	RSS       string           `json:"rss"`
	Updated   time.Time        `json:"updated"`
	Episodes  Episodes         `json:"episodes"`
	Playlists map[string][]int `json:"playlists"`
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

// Filter list by regular expression (see RE2 spec for syntax)
func (f *Feed) Filter(e string) Episodes {
	var (
		out = Episodes{}
		r   = regexp.MustCompile(e)
	)

	for _, ep := range f.Episodes {
		if r.MatchString(ep.Title) {
			out = append(out, ep)
		}
	}

	return out
}

// Set or range of episodes, a range should match a pattern like 0-100
// a set should be a comma separated list (0,34,96) with no spaces
func (f *Feed) Set(s string) Episodes {
	var out Episodes

	switch {
	case strings.Contains(s, ","):
		set := strings.Split(s, ",")
		for _, i := range set {
			id, _ := strconv.Atoi(i)
			if id < len(f.Episodes) {
				out = append(out, f.Episodes[id])
			}
		}

	case strings.Contains(s, "-"):
		var (
			set      = strings.Split(s, "-")
			first, _ = strconv.Atoi(set[0])
			last, _  = strconv.Atoi(set[1])
		)

		if last+1 > len(f.Episodes) {
			last = len(f.Episodes) - 1
		}

		out = f.Episodes[first : last+1]
	}

	return out
}

// Playlist of episodes
func (f *Feed) Playlist(s string) Episodes {
	var out Episodes
	if list, ok := f.Playlists[s]; ok {
		for _, id := range list {
			out = append(out, f.Episodes[id])
		}
	}

	return out
}

// Update the feed
func (f *Feed) Update() error {
	latest, err := FromRSS(f.RSS)
	if err != nil {
		return err
	}

	// Check if the latest publish date is different.
	// Since we sort oldest to newest by default new episodes should only appear at the end
	//if latest.Updated.After(f.Updated) || len(latest.Episodes) != len(f.Episodes) {
	f.Updated = latest.Updated

	for i, ep := range latest.Episodes {
		if i < len(f.Episodes) {
			f.Episodes[i].ID = ep.ID
			f.Episodes[i].Title = ep.Title
			f.Episodes[i].URL = ep.URL
			f.Episodes[i].Mp3 = ep.Mp3
			f.Episodes[i].Length = ep.Length
			f.Episodes[i].Published = ep.Published

		} else if i >= len(f.Episodes) {
			fmt.Printf("-> New episode: %s\n", ep.Title)
			f.Episodes = append(f.Episodes, ep)
		}
	}
	//}

	return nil
}

// String implements the Stringer interface
func (f *Feed) String() string {
	return fmt.Sprintf("Title:\t%s\nURL:\t%s\nRSS:\t%s\nUpdated:\t%s\nEpisodes:\t%d/%d\nPlaylists:\t%s\n",
		f.Title, f.URL, f.RSS, f.Updated.Format("2006-01-02"), len(f.Episodes), f.Played(), listKeys(f.Playlists))
}

func listKeys(in map[string][]int) string {
	var k []string
	for key := range in {
		k = append(k, key)
	}
	return strings.Join(k, ", ")
}

// Feed list sortable by most reent update
type Feeds []*Feed

// Implement sort interface by last update for Feeds
func (f Feeds) Len() int           { return len(f) }
func (f Feeds) Less(i, j int) bool { return f[i].Updated.After(f[j].Updated) }
func (f Feeds) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }

// Episode data
type Episode struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	Mp3       string    `json:"mp3"`
	Length    string    `json:"length"`
	Published time.Time `json:"published"`
	Played    bool      `json:"played"`
	Elapsed   int       `json:"elapsed"`
}

// String implements the Stringer interface
func (e *Episode) String() string {
	return fmt.Sprintf("Title:\t%s\nID:\t%d\nURL:\t%s\nMP3:\t%s\nUpdated:\t%s\nPlayed:\t%v\nElapsed:\t%s\n",
		e.Title, e.ID, e.URL, e.Mp3, e.Published.Format("2006-01-02"), e.Played, ParseElapsed(e.Elapsed))
}

// Episodes is its own type in order to implement a sort interface
type Episodes []*Episode

// Implement sort interface by publish date for Episodes
func (e Episodes) Len() int           { return len(e) }
func (e Episodes) Less(i, j int) bool { return e[i].Published.Before(e[j].Published) }
func (e Episodes) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }

// Set episode IDs post-sort
func (e Episodes) setIDs() {
	for i, ep := range e {
		ep.ID = i
	}
}

// FromRSS creates a new Feed obj by parsing data from an rss url
func FromRSS(url string) (Feed, error) {
	var fd Feed

	// Parse feed
	f, err := gofeed.NewParser().ParseURL(url)
	if err != nil {
		return fd, err
	}

	var feedPub *time.Time
	if f.Published != "" {
		feedPub = f.PublishedParsed
	} else if f.Updated != "" {
		feedPub = f.UpdatedParsed
	}

	// Load key data to Feed obj
	fd = Feed{
		Title:   f.Title,
		URL:     f.Link,
		RSS:     url, // Use the passed URL as that will contain auth info if there is any
		Updated: *feedPub,
	}

	for _, item := range f.Items {
		var epPub *time.Time
		if item.Published != "" {
			epPub = item.PublishedParsed
		} else if item.Updated != "" {
			epPub = item.UpdatedParsed
		}

		if len(item.Enclosures) == 0 {
			return fd, fmt.Errorf("invalid feed; no enclosures (i.e mp3 link) found")
		}

		fd.Episodes = append(fd.Episodes, &Episode{
			Title:     item.Title,
			URL:       item.Link,
			Mp3:       item.Enclosures[0].URL,
			Length:    item.Enclosures[0].Length,
			Published: *epPub,
		})
	}

	// Default sort by oldest first
	sort.Sort(fd.Episodes)
	fd.Episodes.setIDs()
	return fd, nil
}

func ParseElapsed(inSeconds int) string {
	minutes := inSeconds / 60
	seconds := inSeconds % 60

	if minutes >= 0 {
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}

	return fmt.Sprintf("%ds", seconds)
}
