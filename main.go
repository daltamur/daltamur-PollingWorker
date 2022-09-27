package main

import (
	"CSC482/polling"
)

func main() {
	var pollingWorker *polling.PollingWorker
	pollingWorker = polling.NewPollingWorker()
	polling.GetRecentArtists(pollingWorker)
}
