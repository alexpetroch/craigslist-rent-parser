package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html"
)

func findNewHomes(url *url.URL, urlPath string, homesReseached map[string]bool, ch chan int) {

	resp, err := http.Get(urlPath)
	check(err)

	defer resp.Body.Close()

	tokenPtr := html.NewTokenizer(resp.Body)
	links := make(map[string]bool)

	for {
		tokenType := tokenPtr.Next()

		if tokenType == html.ErrorToken {
			break
		}

		if tokenType == html.StartTagToken {
			t := tokenPtr.Token()

			isAnchor := t.Data == "a"
			if isAnchor {

				for _, a := range t.Attr {
					if a.Key == "href" {

						if strings.Contains(a.Val, ".html") {
							if !links[a.Val] {
								links[a.Val] = true
							}
						}
						break
					}
				}
			}
		}
	}

	values := url.Query()
	name := values.Get("query")

	f, err := os.Create(name + ".txt")
	check(err)
	defer f.Close()

	newHomes := 0
	for val := range links {

		urlPath = url.Scheme + "://" + url.Host + val
		// if value represents url update path to use it
		if strings.Contains(val, "http") {
			urlPath = val
		}

		if !homesReseached[urlPath] {
			f.WriteString(urlPath)
			f.WriteString("\n")
			newHomes++
		}
	}

	ch <- newHomes
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {

	bytes, err := ioutil.ReadFile("urls.txt")
	check(err)

	urls := strings.Split(string(bytes), "\n")
	newHomes := 0

	bytes, err = ioutil.ReadFile("reseach.csv")
	check(err)

	lines := strings.Split(string(bytes), "\n")
	homesReseached := make(map[string]bool)
	for i, line := range lines {
		// line 0 is reserved. do not use
		if i == 1 {
			parts := strings.Split(line, ",")
			if len(parts) > 0 {
				homesReseached[strings.TrimSpace(parts[0])] = true
			}
		}
	}

	ch := make(chan int)
	for _, urlPath := range urls {

		url, err := url.Parse(urlPath)
		check(err)

		go findNewHomes(url, urlPath, homesReseached, ch)
	}

	newHomes = 0
	for i := 0; i < len(urls); {
		select {
		case count := <-ch:
			newHomes += count
			i++
		}
	}

	fmt.Println("Done. Found new homes", newHomes)
}
