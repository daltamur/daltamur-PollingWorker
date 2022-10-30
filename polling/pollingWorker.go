package polling

import (
	"PollingWorker/artistImageWebscraper"
	"PollingWorker/dataStructs"
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/jamespearly/loggly"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// PollingWorker only has a reference to the loggly client we will use to send log data
type PollingWorker struct {
	logglyClient *loggly.ClientType
}

//this function just instantiates the polling worker and assigns it a loggly client
func NewPollingWorker() *PollingWorker {
	var worker = PollingWorker{logglyClient: nil}
	instantiateLogglyClient(&worker)
	return &worker
}

//to instantiate the loggly client
func instantiateLogglyClient(worker *PollingWorker) {
	var tag string
	tag = "Lastfm-Poll-Worker"
	// Instantiate the client
	worker.logglyClient = loggly.New(tag)
}

//this is the meat and potatoes function of our polling worker
func handleRecentArtistsCall(jsonBody *[]byte) {
	//instantiate the reference to the DB
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create DynamoDB client
	svc := dynamodb.New(sess)

	//call the Getrecenttracks function from the dataStructs package

	var tracks = dataStructs.Getrecenttracks(jsonBody)

	//instantiate a wait group for parallelism
	var wg sync.WaitGroup

	//go through each track struct that is returned
	for index, element := range tracks.MostRecentTrackInfo.Tracks {
		//add 1 to the waitgroup
		wg.Add(1)
		//make placeholder variables bc using the ones from the for loop could be a problem
		var curElement = element
		var curIndex = index
		index := index

		//check if the current track is equal to the mostRecentSongJson. If it is, break out of the loop, no need to waste
		//resources on songs we are not adding to the database
		_, err := os.Stat("mostRecentSong.json")
		var mostRecentSong dataStructs.Track
		if os.IsExist(err) {
			plan, _ := ioutil.ReadFile("mostRecentSong.json")
			err := json.Unmarshal(plan, &mostRecentSong)
			if err != nil {
				panic(err)
			}
			if element.Date.Uts == mostRecentSong.Date.Uts && element.Name == mostRecentSong.Name {
				break
			}
		}

		//run the body of this function in parallel
		go func() {
			//tell the wait group the function is done when the rest of the body finishes
			defer wg.Done()
			//get the artist images using the track data & webscraping
			var artistImages = *artistImageWebscraper.GetArtistImage(&strings.Split(curElement.Url, "/_")[0])
			//if the webscraper found artist images, assign them to the artist struct found in the current track struct
			if len(artistImages) != 0 {
				var smallImage = dataStructs.Image{Size: "small", Text: artistImages[0]}
				var mediumImage = dataStructs.Image{Size: "medium", Text: artistImages[1]}
				var largeImage = dataStructs.Image{Size: "large", Text: artistImages[2]}
				tracks.MostRecentTrackInfo.Tracks[curIndex].Artist.Image = append(tracks.MostRecentTrackInfo.Tracks[curIndex].Artist.Image, smallImage)
				tracks.MostRecentTrackInfo.Tracks[curIndex].Artist.Image = append(tracks.MostRecentTrackInfo.Tracks[curIndex].Artist.Image, mediumImage)
				tracks.MostRecentTrackInfo.Tracks[curIndex].Artist.Image = append(tracks.MostRecentTrackInfo.Tracks[curIndex].Artist.Image, largeImage)
				//if the webscraper found no images, fill up the artist's images with placeholder image URLs
			} else {
				var smallImage = dataStructs.Image{Size: "small", Text: "https://lastfm.freetls.fastly.net/i/u/34s/2a96cbd8b46e442fc41c2b86b821562f.png"}
				var mediumImage = dataStructs.Image{Size: "medium", Text: "https://lastfm.freetls.fastly.net/i/u/64s/2a96cbd8b46e442fc41c2b86b821562f.png"}
				var largeImage = dataStructs.Image{Size: "large", Text: "https://lastfm.freetls.fastly.net/i/u/300x300/2a96cbd8b46e442fc41c2b86b821562f.png"}
				tracks.MostRecentTrackInfo.Tracks[index].Artist.Image = append(tracks.MostRecentTrackInfo.Tracks[index].Artist.Image, smallImage)
				tracks.MostRecentTrackInfo.Tracks[index].Artist.Image = append(tracks.MostRecentTrackInfo.Tracks[index].Artist.Image, mediumImage)
				tracks.MostRecentTrackInfo.Tracks[index].Artist.Image = append(tracks.MostRecentTrackInfo.Tracks[index].Artist.Image, largeImage)
			}
		}()
		randTime := rand.Intn(10-3) + 3
		time.Sleep(time.Duration(randTime) * time.Second)
	}
	wg.Wait()

	//just print out each track's information in a readable way, stop when you reach the last known file
	var foundMostRecentValue = false
	//var mostRecentSongIndex = len(tracks.MostRecentTrackInfo.Tracks)
	var newMostRecentSong dataStructs.Track
	plan, _ := ioutil.ReadFile("mostRecentSong.json")
	_ = json.Unmarshal(plan, &newMostRecentSong)

	for _, element := range tracks.MostRecentTrackInfo.Tracks {
		var mostRecentSong dataStructs.Track
		if element.Date.Text != "" {
			_, err := os.Stat("mostRecentSong.json")
			//no most recent song has ever been created, so we just set get that set up
			if !os.IsNotExist(err) {
				plan, _ := ioutil.ReadFile("mostRecentSong.json")
				err := json.Unmarshal(plan, &mostRecentSong)
				if err != nil {
					panic(err)
				}
				if mostRecentSong.Date.Uts == element.Date.Uts {
					foundMostRecentValue = true
					//mostRecentSongIndex = index
				}
			}
			if foundMostRecentValue {
				fmt.Println("No more new songs to add to database right now!")
				break
			}

			//convert uts to EST
			location, _ := time.LoadLocation("America/New_York")
			i, _ := strconv.ParseInt(element.Date.Uts, 10, 64)
			timeVal := time.Unix(i, 0)
			curTime := timeVal.In(location).Format(time.Layout)
			curTimeArr := strings.Split(curTime, " ")
			finalCurTime := curTimeArr[1]
			DateValue := curTimeArr[0] + "/" + curTimeArr[2][1:]
			//UTCValue := element.Date.Uts
			ArtistValue := element.Artist.Text
			ArtistImg := element.Artist.Image
			Trackval := element.Name
			Trackimg := element.Image
			Albumv := element.Album.Text
			if Albumv == "" {
				Albumv = "Single"
			}
			ESTTimeVal := finalCurTime
			ThisTrack := element
			go func() {
				//hash ThisTrack
				trackbits := new(bytes.Buffer)
				err := json.NewEncoder(trackbits).Encode(ThisTrack)
				hashVal := md5.Sum(trackbits.Bytes())
				value, _ := strconv.Atoi(ThisTrack.Date.Uts)
				track := dataStructs.DBTrack{
					Hashval:        fmt.Sprintf("%x", hashVal),
					DateVal:        DateValue,
					UTSVal:         value,
					ArtistVal:      ArtistValue,
					ArtistValImage: ArtistImg,
					TrackVal:       Trackval,
					TrackValImage:  Trackimg,
					AlbumVal:       Albumv,
					ESTTime:        ESTTimeVal,
				}

				//marshall struct into dynamoattribute
				av, err := dynamodbattribute.MarshalMap(track)

				//create instance of Tracks DB
				tableName := "daltamur-LastFMTracks"
				input := &dynamodb.PutItemInput{
					Item:      av,
					TableName: aws.String(tableName),
				}

				_, err = svc.PutItem(input)
				if err != nil {
					log.Fatalf("Got error calling PutItem: %s", err)
				}
			}()

		}

	}

	foundMostRecentValue = false

	for index, element := range tracks.MostRecentTrackInfo.Tracks {
		var mostRecentSong dataStructs.Track
		if element.Date.Text != "" {
			_, err := os.Stat("mostRecentSong.json")
			//no most recent song has ever been created, so we just set get that set up
			if os.IsNotExist(err) {
				file, _ := json.MarshalIndent(element, "", " ")
				_ = ioutil.WriteFile("mostRecentSong.json", file, 0644)

			} else {
				plan, _ := ioutil.ReadFile("mostRecentSong.json")
				err := json.Unmarshal(plan, &mostRecentSong)
				if err != nil {
					panic(err)
				}
				if mostRecentSong.Date.Uts == element.Date.Uts {
					foundMostRecentValue = true
					//mostRecentSongIndex = index
				}
			}
			if foundMostRecentValue {
				fmt.Println("No more new songs to report right now!")
				break
			}
			println("Song ", index)
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
			//var artistImages = *artistImageWebscraper.GetArtistImage(&strings.Split(element.Url, "/_")[0])
			//if len(artistImages) != 0 {
			//	var smallImage = dataStructs.Image{Size: "small", Text: artistImages[0]}
			//	var mediumImage = dataStructs.Image{Size: "medium", Text: artistImages[1]}
			//	var largeImage = dataStructs.Image{Size: "large", Text: artistImages[2]}
			//	tracks.MostRecentTrackInfo.Tracks[index].Artist.Image = append(tracks.MostRecentTrackInfo.Tracks[index].Artist.Image, smallImage)
			//	tracks.MostRecentTrackInfo.Tracks[index].Artist.Image = append(tracks.MostRecentTrackInfo.Tracks[index].Artist.Image, mediumImage)
			//	tracks.MostRecentTrackInfo.Tracks[index].Artist.Image = append(tracks.MostRecentTrackInfo.Tracks[index].Artist.Image, largeImage)
			//} else {
			//	var smallImage = dataStructs.Image{Size: "small", Text: "https://lastfm.freetls.fastly.net/i/u/34s/2a96cbd8b46e442fc41c2b86b821562f.png"}
			//	var mediumImage = dataStructs.Image{Size: "medium", Text: "https://lastfm.freetls.fastly.net/i/u/64s/2a96cbd8b46e442fc41c2b86b821562f.png"}
			//	var largeImage = dataStructs.Image{Size: "large", Text: "https://lastfm.freetls.fastly.net/i/u/300x300/2a96cbd8b46e442fc41c2b86b821562f.png"}
			//	tracks.MostRecentTrackInfo.Tracks[index].Artist.Image = append(tracks.MostRecentTrackInfo.Tracks[index].Artist.Image, smallImage)
			//	tracks.MostRecentTrackInfo.Tracks[index].Artist.Image = append(tracks.MostRecentTrackInfo.Tracks[index].Artist.Image, mediumImage)
			//	tracks.MostRecentTrackInfo.Tracks[index].Artist.Image = append(tracks.MostRecentTrackInfo.Tracks[index].Artist.Image, largeImage)
			//}
			println("Artist Image Links: ")
			for _, curImage := range element.Artist.Image {
				println(curImage.Size, ": ", curImage.Text)
			}
			println()
			println("Album Art Links: ")
			for _, curImage := range element.Image {
				println(curImage.Size, ": ", curImage.Text)
			}
			println("------------------------------------------------------------------------")
		}
	}
	//rewrite the most recent song as the topmost song, so long as it is not the "now playing" song
	var curMostRecentSong = tracks.MostRecentTrackInfo.Tracks[0]
	if curMostRecentSong.Date.Text != "" {
		file, _ := json.MarshalIndent(curMostRecentSong, "", " ")
		_ = ioutil.WriteFile("mostRecentSong.json", file, 0644)
	} else {
		curMostRecentSong = tracks.MostRecentTrackInfo.Tracks[1]
		file, _ := json.MarshalIndent(curMostRecentSong, "", " ")
		_ = ioutil.WriteFile("mostRecentSong.json", file, 0644)
	}

	println()
	println()
	println("User: ", tracks.MostRecentTrackInfo.Attr.User)
	println("Overall number of tracks listened to: ", tracks.MostRecentTrackInfo.Attr.Total)
	println()
}

func GetRecentArtists(worker *PollingWorker) {
	//make the request to the LastFM api
	token := os.Getenv("LASTFM_TOKEN")
	if token == "" {
		fmt.Println("ERROR: No LAST.FM Token detected!")
		os.Exit(-1)
	}
	//start a timer, so we can see how long it takes to get a response
	//this for loop is solely so we can add multiple tables if needed
	for page := 1; page < 2; page++ {
		start := time.Now()
		response, err := http.Get("https://ws.audioscrobbler.com/2.0/?method=user.getrecenttracks&limit=200&&user=witless_wisdom&api_key=" + token + "&format=json&page=" + strconv.Itoa(page))
		if err != nil {
			var errorMessage = err.Error()
			doLogglyMessage(worker, "GET Request error: "+errorMessage+"\n"+"No data gathered", "error")
		} else {
			//close the response reader when the rest of the function finishes
			defer func(Body io.ReadCloser) {
				_ = Body.Close()
			}(response.Body)

			//save the amount of time it took to get a response to a variable to print out later, make sure to use
			var requestTime = time.Since(start)
			var requestTimeString = requestTime.String()
			requestTimeString = strings.Replace(requestTimeString, "s", " seconds", 1)
			requestTimeString = strings.Replace(requestTimeString, "m", " milliseconds", 1)

			//print the repsonse's status code
			fmt.Println("Response Status: ", response.Status)
			jsonDataFromHTTP, _ := ioutil.ReadAll(response.Body)
			var responseSize = float32(len(jsonDataFromHTTP)) / 1000.0

			//start a new timer to see how long it takes to process the response
			start = time.Now()
			handleRecentArtistsCall(&jsonDataFromHTTP)
			var totalTime = time.Since(start)
			var totalTimeString = totalTime.String()
			totalTimeString = strings.Replace(totalTimeString, "s", " seconds", 1)
			totalTimeString = strings.Replace(totalTimeString, "m", " milliseconds", 1)

			//send loggly the size of the response and how long it took to get it and then how long it took to process it
			doLogglyMessage(worker, "GET Request accepted\nSize of response is "+fmt.Sprintf("%f", responseSize)+" kilobytes.\nResponse took "+requestTimeString+" to receive and "+totalTimeString+" to process", "info")
			fmt.Println("Cur page: ", page)
		}
	}
}

func doLogglyMessage(worker *PollingWorker, message string, level string) {
	err := worker.logglyClient.EchoSend(level, message)
	if err != nil {
		fmt.Println(err)
	}
}
