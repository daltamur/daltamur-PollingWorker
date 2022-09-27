package artistImageWebscraper

import (
	"strings"
	"time"
)
import (
	"github.com/gocolly/colly"
)

func GetArtistImage(artistUrl *string) *[]string {
	var imageAddress = new(string)
	var imageAddresses []string
	c := colly.NewCollector()
	c.SetRequestTimeout(120 * time.Second)

	c.OnRequest(func(r *colly.Request) {})

	c.OnResponse(func(r *colly.Response) {})

	c.OnError(func(r *colly.Response, e error) {
		//fmt.Println("Got this error:", e)
	})

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

	c.Visit(strings.Split(*artistUrl, "/_")[0])
	return &imageAddresses

}
