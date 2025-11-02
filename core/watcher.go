package core

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type Watched struct {
	Hash       string        `json:"hash"`
	Body       *html.Node    `json:"body"`
	LastUpdate time.Time     `json:"last_update"`
	Interval   int64         `json:"interval"`
	Request    *WatchRequest `json:"request"`
}

type WatcherService struct {
	WatchedUrls      []*Watched
	httpService      *HttpService
	diffService      *DiffService
	OnChangeCallback func(watched *WatchResponse)
}

func NewWatcherService() *WatcherService {
	watcher := new(WatcherService)
	watcher.WatchedUrls = nil
	watcher.httpService = NewHttpService()
	watcher.diffService = NewDiffService()
	return watcher
}

func (watcherService *WatcherService) Start() {
	for {
		var wg sync.WaitGroup
		for _, watched := range watcherService.WatchedUrls {
			current := time.Now().Unix()
			nextUpdate := watched.LastUpdate.Unix() + watched.Interval
			if nextUpdate < current {
				wg.Add(1)
				go watcherService.Request(&wg, watched)
			}
		}
		wg.Wait()
	}
}

func (watcherService *WatcherService) Request(wg *sync.WaitGroup, watched *Watched) {
	defer wg.Done()
	newBody, _ := watcherService.httpService.GetBody(watched.Request.URL.String())
	newHash := watcherService.httpService.GetNodeHash(newBody)

	// Check if this is not the first check and hash has changed
	if watched.Hash != "" && watched.Hash != newHash {
		var response = new(WatchResponse)
		response.URL = watched.Request.URL
		response.ChannelID = watched.Request.FeedbackChannelInfo.ChannelID
		response.Diff = watcherService.diffService.Diff(watched.Body, newBody)
		fmt.Printf("Change detected for %s\n", watched.Request.URL)
		fmt.Printf("%s\n", response.Diff)

		// Trigger callback if set
		if watcherService.OnChangeCallback != nil {
			watcherService.OnChangeCallback(response)
		}
	}

	watched.Body = newBody
	watched.Hash = newHash
	watched.LastUpdate = time.Now()

	fmt.Println(watched.Request.URL)
	watched.LastUpdate = time.Now()
}

func (watcherService *WatcherService) Register(request *WatchRequest) error {
	watched := new(Watched)
	watched.Request = request
	watched.Interval = 10

	watched.Body, _ = watcherService.httpService.GetBody(watched.Request.URL.String())
	watched.Hash = watcherService.httpService.GetNodeHash(watched.Body)

	watched.LastUpdate = time.Now()

	fmt.Println(watched.Request.URL)
	watcherService.Insert(watched)
	return nil
}

func (watcherService *WatcherService) Insert(watch *Watched) {
	watcherService.persistWatched(watch)
	isAfterIndex := func(i int) bool { return watcherService.WatchedUrls[i].LastUpdate.After(watch.LastUpdate) }
	idx := sort.Search(len(watcherService.WatchedUrls), isAfterIndex)
	watcherService.WatchedUrls = append(watcherService.WatchedUrls, nil)
	copy(watcherService.WatchedUrls[idx+1:], watcherService.WatchedUrls[idx:])
	watcherService.WatchedUrls[idx] = watch
}

func (watcherService *WatcherService) GetWatched() []*Watched {
	return watcherService.WatchedUrls
}

func (watcherService *WatcherService) persistWatched(watch *Watched) {
	// watcherService.repository
}
