package main

import (
	"PollingWorker/polling"
	"math/rand"
	"time"
)

func main() {
	//create a polling worker instance
	var pollingWorker *polling.PollingWorker
	pollingWorker = polling.NewPollingWorker()
	polling.GetRecentArtists(pollingWorker)
	println()
	println()

	//set the ticker to run in the background and tick every ten seconds
	go func() {
		randTime := rand.Intn(240-180) + 180
		ticker := time.NewTicker(time.Duration(randTime) * time.Minute)
		for _ = range ticker.C {
			//run the method that will run a get request from the API
			//turn the JSON into a struct
			//and then print out the struct values.
			polling.GetRecentArtists(pollingWorker)
			println()
			println()
		}
	}()
	select {}
}
