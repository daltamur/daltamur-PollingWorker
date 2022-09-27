package polling

import (
	"CSC482/artistImageWebscraper"
	"CSC482/dataStructs"
	"encoding/json"
	"fmt"
	"github.com/jamespearly/loggly"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

type PollingWorker struct {
	logglyClient *loggly.ClientType
}

func NewPollingWorker() *PollingWorker {
	var worker = PollingWorker{logglyClient: nil}
	instantiateLogglyClient(&worker)
	return &worker
}

func instantiateLogglyClient(worker *PollingWorker) {
	var tag string
	tag = "Lastfm-Poll-Worker"
	// Instantiate the client
	worker.logglyClient = loggly.New(tag)
}

func handleRecentArtistsCall(jsonBody *[]byte, worker *PollingWorker) {
	var tracks = dataStructs.Getrecenttracks(jsonBody)
	errUnmarshal := json.Unmarshal(*jsonBody, &tracks)
	if errUnmarshal != nil {
		panic(errUnmarshal)
	}

	var wg sync.WaitGroup
	for index, element := range tracks.MostRecentTrackInfo.Tracks {
		wg.Add(1)
		var curElement = element
		var curIndex = index
		go func() {
			defer wg.Done()
			var artistImages = *artistImageWebscraper.GetArtistImage(&strings.Split(curElement.Url, "/_")[0])
			if len(artistImages) != 0 {
				var smallImage = dataStructs.Image{Size: "small", Text: artistImages[0]}
				var mediumImage = dataStructs.Image{Size: "medium", Text: artistImages[1]}
				var largeImage = dataStructs.Image{Size: "large", Text: artistImages[2]}
				tracks.MostRecentTrackInfo.Tracks[curIndex].Artist.Image = append(tracks.MostRecentTrackInfo.Tracks[curIndex].Artist.Image, smallImage)
				tracks.MostRecentTrackInfo.Tracks[curIndex].Artist.Image = append(tracks.MostRecentTrackInfo.Tracks[curIndex].Artist.Image, mediumImage)
				tracks.MostRecentTrackInfo.Tracks[curIndex].Artist.Image = append(tracks.MostRecentTrackInfo.Tracks[curIndex].Artist.Image, largeImage)
			}
		}()
	}
	wg.Wait()

	for index, element := range tracks.MostRecentTrackInfo.Tracks {
		println("Song ", index+1)
		if element.Date.Text != "" {
			println("Played ", element.Name, " (MBID: ", element.Mbid, ")")
		} else {
			println("Now playing ", element.Name, " (MBID: ", element.Mbid, ")")
		}
		println("By ", element.Artist.Text, " (MBID: ", element.Artist.Mbid, ")")
		println("From the album ", element.Album.Text, " (MBID: ", element.Album.Mbid, ")")
		if element.Date.Text != "" {
			println("At ", element.Date.Text, " (UTS Time: ", element.Date.Uts, ")")
		}
		if element.Streamable == "0" {
			println("Track is not streamable")
		} else {
			println("Track is streamable")
		}
		println("Learn more about the track at ", element.Url)
		println()
		//element.Artist.Image = *artistImageWebscraper.GetArtistImage(&strings.Split(element.Url, "/_")[0], &element.Artist.Image)
		println("Artist Image Links: ")
		for _, curImage := range element.Artist.Image {
			println(curImage.Size, ": ", curImage.Text)
		}
		println()
		println("Album Art Links: ")
		for _, curImage := range element.Image {
			println(curImage.Size, ": ", curImage.Text)
		}
		println()
		//somewhere here we need the polling worker to do somethin
	}
}

func GetRecentArtists(worker *PollingWorker) {
	//make the request to the LastFM api
	response, err := http.Get("https://ws.audioscrobbler.com/2.0/?method=user.getrecenttracks&limit=100&user=witless_wisdom&api_key=966b5110aac66a9e98458f93a8362619&format=json")
	if err != nil {
		var errorMessage = err.Error()
		doLogglyMessage(worker, "GET Request error: "+errorMessage+"\n"+"No data gathered", "error")
	} else {
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(response.Body)
		fmt.Println("Response Status: ", response.Status)
		jsonDataFromHTTP, errReadJson := ioutil.ReadAll(response.Body)
		var responseSize = len(jsonDataFromHTTP)
		if errReadJson != nil {
			doLogglyMessage(worker, "GET Request accepted\nSize of response is "+strconv.Itoa(responseSize)+" bytes.", "info")
			panic(errReadJson)
		}

		handleRecentArtistsCall(&jsonDataFromHTTP, worker)
		doLogglyMessage(worker, "GET Request accepted\nSize of response is "+strconv.Itoa(responseSize)+" bytes.", "info")
	}
}

func doLogglyMessage(worker *PollingWorker, message string, level string) {

	// Valid EchoSend (message echoed to console and no error returned)
	//err := worker.logglyClient.EchoSend("info", "Good morning!")
	//fmt.Println("err:", err)

	//// Valid Send (no error returned)
	err := worker.logglyClient.EchoSend(level, message)
	if err != nil {
		fmt.Println(err)
	}
	//
	//// Invalid EchoSend -- message level error
	//err = worker.logglyClient.EchoSend("blah", "blah")
	//fmt.Println("err:", err)

}
