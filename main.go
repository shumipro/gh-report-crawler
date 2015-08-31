package main

import "os"
import "os/signal"
import "time"
import "fmt"
import "sync"
import "net/http"

import "golang.org/x/oauth2"
import "github.com/google/go-github/github"

type Transport struct {
	lastEtag  string
	Transport http.RoundTripper
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(t.lastEtag) > 0 {
		req.Header.Set("If-None-Match", t.lastEtag)
	}
	resp, err := t.Transport.RoundTrip(req)
	t.lastEtag = resp.Header.Get("Etag")
	return resp, err
}

func startCrawler(user string, accessToken string, quit chan struct{}, wg *sync.WaitGroup) {
	authTransport := &oauth2.Transport{
		Source: oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken}),
	}
	cacheTransport := &Transport{Transport: authTransport, lastEtag: ""}
	httpClient := &http.Client{Transport: cacheTransport}
	client := github.NewClient(httpClient)

L:
	for {
		select {
		case <-quit:
			// TODO: cleanup
			fmt.Fprintln(os.Stdout, "Clean up")
			wg.Done()
			break L
		default:
			fmt.Fprintln(os.Stdout, "Call API")
			events, resp, err := client.Activity.ListEventsPerformedByUser(user, false, &github.ListOptions{PerPage: 300})
			if err != nil && resp.StatusCode != 304 {
				fmt.Fprintln(os.Stderr, err)
			}
			fmt.Fprintln(os.Stdout, resp.StatusCode)
			fmt.Fprintln(os.Stdout, events)
		}
		time.Sleep(3 * time.Second)
	}
}

func startCrawlers(done chan struct{}, quit chan struct{}) {
	wg := &sync.WaitGroup{}
	// TODO: fetch username and accessToken
	wg.Add(1)
	go startCrawler("kitak", "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX", quit, wg)
	wg.Wait()
	done <- struct{}{}
}

func main() {
	done := make(chan struct{})
	quit := make(chan struct{})
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		<-sig
		close(quit)
	}()

	go startCrawlers(done, quit)

	<-done
}
