package dataStructs

import "encoding/json"

type Artist struct {
	Mbid  string `json:"mbid"`
	Image []Image
	Text  string `json:"#text"`
}

type Album struct {
	Mbid string `json:"mbid"`
	Text string `json:"#text"`
}

type Date struct {
	Uts  string `json:"uts"`
	Text string `json:"#text"`
}

type Image struct {
	Size string `json:"size"`
	Text string `json:"#text"`
}

type Track struct {
	Artist     Artist  `json:"artist"`
	Streamable string  `json:"streamable"`
	Mbid       string  `json:"mbid"`
	Album      Album   `json:"album"`
	Name       string  `json:"name"`
	Url        string  `json:"url"`
	Date       Date    `json:"date"`
	Image      []Image `json:"image"`
}

type Attr struct {
	User       string `json:"user"`
	TotalPages string `json:"totalPages"`
	Page       string `json:"page"`
	PerPage    string `json:"perPage"`
	Total      string `json:"total"`
}

type TracksAndAttr struct {
	Tracks []Track `json:"track"`
	Attr   Attr    `json:"@attr"`
}

type RecentTracks struct {
	MostRecentTrackInfo TracksAndAttr `json:"recenttracks"`
}

func Getrecenttracks(inputJson *[]byte) *RecentTracks {
	var tracks RecentTracks
	err := json.Unmarshal([]byte(*inputJson), &tracks)
	if err != nil {
		panic(err)
	}
	return &tracks
}
