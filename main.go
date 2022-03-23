package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/gocolly/colly"
)

type star struct {
	Name      string
	Photo     string
	JobTitle  string
	BirthDate string
	Bio       string
	TopMovies []movie
}

type movie struct {
	Title string
	Year  string
}

func main() {
	//pointer are used so to be used from the commmand line
	month := flag.Int("month", 1, "Month to fetch birthdays for ")
	day := flag.Int("day", 1, "Day to fetch birthday for")
	flag.Parse()
	crawl(*month, *day)
}

func crawl(month int, day int) {
	// c is used to crawl through the data that is received every page
	// it just goes over profiles
	c := colly.NewCollector(
		colly.AllowedDomains("imdb.com", "www.imdb.com"),
	)

	// infocollector is used to move inside each profile
	infoCollector := c.Clone()

	c.OnHTML(".mode-detail", func(e *colly.HTMLElement) {
		profileUrl := e.ChildAttr("div.lister-item-image > a", "href")
		profileUrl = e.Request.AbsoluteURL(profileUrl)
		// infoCollector goes into each profile and collects data
		infoCollector.Visit(profileUrl)
	})

	// movies through pages --- a.lister-page-next is imdb specific
	c.OnHTML("a.lister-page-next", func(e *colly.HTMLElement) {
		nextPage := e.Request.AbsoluteURL(e.Attr("href"))
		c.Visit(nextPage)
	})

	//infoCollector will makes sence of ids in apges and extract data into struct
	infoCollector.OnHTML("#content-2-wide", func(e *colly.HTMLElement) {
		tempProfile := star{}

		tempProfile.Name = e.ChildText("h1.header > span.itemprop")
		tempProfile.Photo = e.ChildAttr("#name-poster", "src")
		tempProfile.JobTitle = e.ChildText("#name-job-categories > a > span.itemprop")
		tempProfile.BirthDate = e.ChildAttr("#name-born-info time", "datetime")
		tempProfile.Bio = strings.TrimSpace(e.ChildText("#name-bio-text > div.name-trivia-bio-text > div.inline"))

		e.ForEach("div.knownfor-title", func(_ int, kf *colly.HTMLElement) {
			tempMovie := movie{}

			tempMovie.Title = kf.ChildText("div.knownfor-title-role > a.knownfor-ellipsis")
			tempMovie.Year = kf.ChildText("div.knownfor-year > span.knownfor.ellipsis")
			// appending movies to topmovies slice
			tempProfile.TopMovies = append(tempProfile.TopMovies, tempMovie)
		})
		js, err := json.MarshalIndent(tempProfile, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(js))

	})
	// to let us know that c is visiting some url
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting: ", r.URL.String())
	})
	// to let knwo that infocollector is visiting some profile
	infoCollector.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting profile url: ", r.URL.String())
	})

	// this is the url we are supposed to crawl
	startUrl := fmt.Sprintf("https://www.imdb.com/search/name/?birth_monthday=%d-%d", month, day)
	c.Visit(startUrl)
}
