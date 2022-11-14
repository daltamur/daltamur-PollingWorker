package artistImageWebscraper

import (
	"PollingWorker/polling"
	"net/url"
	"strings"
	"time"
)
import (
	"github.com/gocolly/colly"
)

var proxies = []*url.URL{
	{Host: "127.0.0.1:8080"},
	{Host: "127.0.0.1:8081"},
}

func GetArtistImage(artistUrl *string) *[]string {

	//create vars to hold our imageURLs
	var imageAddress = new(string)
	var imageAddresses []string

	//instantiate the web scraper
	c := colly.NewCollector()
	c.CheckHead = true

	//timeout after 120 seconds if nothing comes back
	c.SetRequestTimeout(120 * time.Second)

	/*
		do this when we get an html response
		to get the url of the artist image we need to look at the div[class=header-new-gallery-outer] tag and
		then the header-new-gallery header-new-gallery--link hidden-xs Child attribute to get a specific code that we need to
		append to our image urls to get the artist image
	*/
	c.OnHTML("div[class=header-new-gallery-outer]", func(e *colly.HTMLElement) {
		if e.ChildAttr("a", "class") == "header-new-gallery\n                            header-new-gallery--link\n                            hidden-xs\n                            link-block-target" {
			*imageAddress = e.ChildAttr("a", "href")
			positionOfID := strings.LastIndex(*imageAddress, "/") + 1

			//small image
			imageAddresses = append(imageAddresses, "https://lastfm.freetls.fastly.net/i/u/174s/"+(*imageAddress)[positionOfID:]+".jpg")

			//medium image
			imageAddresses = append(imageAddresses, "https://lastfm.freetls.fastly.net/i/u/300x300/"+(*imageAddress)[positionOfID:]+".jpg")

			//large image
			imageAddresses = append(imageAddresses, "https://lastfm.freetls.fastly.net/i/u/470x470/"+(*imageAddress)[positionOfID:]+".jpg")
		}
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("referer", "htt[s://www.google.com")
		r.Headers.Set("pragma", "no-cache")
		r.Headers.Set("dnt", "1")
		r.Headers.Set("upgradable-iunscure-requests", "1")
	})

	//Our image URL comes in like last.fm/<Some Artist>/_/<Some Song>
	//We only need last.fm/<some_artist> to get our address we will webscrape.
	err := c.Visit(strings.Split(*artistUrl, "/_")[0])
	for err != nil {
		//for some reason in docker it will fail sometimes (this never happens when I run just the compiled binaries like normal).
		//just keep trying until we don't fail anymore
		stringVal := "Failed to visit " + *artistUrl + " trying again..."
		polling.DoLogglyMessage(polling.NewPollingWorker(), stringVal, "info")
		time.Sleep(1 * time.Minute)
		err = c.Visit(strings.Split(*artistUrl, "/_")[0])
	}

	//return the address of the imageAddresses slice that hold all the artist's images in various sizes
	return &imageAddresses

}
