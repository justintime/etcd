package store

import (
	"path"
	"strconv"
	"strings"
)

type WatcherHub struct {
	watchers map[string][]Watcher
}

type Watcher struct {
	c     chan Response
}

// global watcher
var w *WatcherHub

// init the global watcher
func init() {
	w = createWatcherHub()
}

// create a new watcher
func createWatcherHub() *WatcherHub {
	w := new(WatcherHub)
	w.watchers = make(map[string][]Watcher)
	return w
}

func GetWatcherHub() *WatcherHub {
	return w
}

// register a function with channel and prefix to the watcher
func AddWatcher(prefix string, c chan Response, sinceIndex uint64) error {

	prefix = "/" + path.Clean(prefix)

	if sinceIndex != 0 && sinceIndex >= s.ResponseStartIndex {
		for i := sinceIndex; i <= s.Index; i++ {
			if check(prefix, i) {
				c <- s.ResponseMap[strconv.FormatUint(i, 10)]
				return nil
			}
		}
	}

	_, ok := w.watchers[prefix]

	if !ok {

		w.watchers[prefix] = make([]Watcher, 0)

		watcher := Watcher{c}

		w.watchers[prefix] = append(w.watchers[prefix], watcher)
	} else {

		watcher := Watcher{c}

		w.watchers[prefix] = append(w.watchers[prefix], watcher)
	}

	return nil
}

// check if the response has what we are watching
func check(prefix string, index uint64) bool {

	resp, ok := s.ResponseMap[strconv.FormatUint(index, 10)]

	if !ok {
		// not storage system command
		return false
	} else {
		path := resp.Key
		if strings.HasPrefix(path, prefix) {
			prefixLen := len(prefix)
			if len(path) == prefixLen || path[prefixLen] == '/' {
				return true
			}

		}
	}

	return false
}


// notify the watcher a action happened
func notify(resp Response) error {
	resp.Key = path.Clean(resp.Key)

	segments := strings.Split(resp.Key, "/")
	currPath := "/"

	// walk through all the pathes
	for _, segment := range segments {
		currPath = path.Join(currPath, segment)

		watchers, ok := w.watchers[currPath]

		if ok {

			newWatchers := make([]Watcher, 0)
			// notify all the watchers
			for _, watcher := range watchers {
				watcher.c <- resp
			}

			if len(newWatchers) == 0 {
				// we have notified all the watchers at this path
				// delete the map
				delete(w.watchers, currPath)
			} else {
				w.watchers[currPath] = newWatchers
			}
		}

	}

	return nil
}