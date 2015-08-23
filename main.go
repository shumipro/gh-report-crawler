package main

import "os"
import "os/signal"
import "time"
import "fmt"

import "golang.org/x/oauth2"
import "github.com/google/go-github/github"

func startCrawler(user string, accessToken string) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)
	client := github.NewClient(tc)

	for {
		client.Activity.ListEventsPerformedByUser(user, false, &github.ListOptions{PerPage: 300})
		time.Sleep(3 * time.Second)
	}
}

func startCrawlers(done chan struct{}, quit chan struct{}) {
	// TODO: fetch username and accessToken

	users := []int{1, 2, 3}
	for _, user := range users {
		fmt.Println(user)
		//go startCrawler(user.Name, user.AccessToken)
	}

	for {
		select {
		case <-quit:
			done <- struct{}{}
		}
	}
}

func main() {
	done := make(chan struct{})
	quit := make(chan struct{})
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	go func() {
		<-sig
		quit <- struct{}{}
	}()

	go startCrawlers(done, quit)

	<-done
}
