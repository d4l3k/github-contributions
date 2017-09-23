package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/pkg/errors"
)

const format = "2006-01-02"

var (
	user = flag.String("user", "d4l3k", "the user to fetch contributions for")

	templates = []string{
		"https://github.com/users/d4l3k/created_issues?from=%s&to=%s",
		"https://github.com/users/d4l3k/created_commits?from=%s&to=%s",
		"https://github.com/users/d4l3k/created_pull_requests?from=%s&to=%s",
		"https://github.com/users/d4l3k/created_pull_request_reviews?from=%s&to=%s",
		"https://github.com/users/d4l3k/created_repositories?from=%s&to=%s",
	}
)

var limit = time.NewTicker(1 * time.Second)

func getLinks(link string) ([]string, error) {
	<-limit.C
	log.Printf("Fetching: %s", link)
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("got status code %d: %s", resp.StatusCode, resp.Status)
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, errors.Wrapf(err, "error fetching %q", link)
	}
	base, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	var links []string
	var err2 error
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href := s.AttrOr("href", "")
		rel, err := url.Parse(href)
		if err != nil {
			err2 = err
			return
		}
		resolved := base.ResolveReference(rel)
		resolved.RawQuery = ""
		split := strings.Split(resolved.Path, "/")
		if len(split) > 3 {
			resolved.Path = strings.Join(split[:3], "/")
		}
		links = append(links, resolved.String())
	})
	if err2 != nil {
		return nil, err2
	}
	return links, nil
}

func main() {
	start := time.Date(2011, time.January, 1, 0, 0, 0, 0, time.Local)
	end := time.Now()

	var allLinks []string

	for ; start.Before(end); start = start.Add(28 * 24 * time.Hour) {
		date := start.Format(format)
		dateTo := start.Add(31 * 24 * time.Hour).Format(format)
		for _, tmpl := range templates {
			links, err := getLinks(fmt.Sprintf(tmpl, date, dateTo))
			if err != nil {
				log.Fatal(err)
			}
			log.Println(links)
			allLinks = append(allLinks, links...)
		}
	}
	if err := ioutil.WriteFile("links.txt", []byte(strings.Join(allLinks, "\n")+"\n"), 0755); err != nil {
		log.Fatal(err)
	}
}
