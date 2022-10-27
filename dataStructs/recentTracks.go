package dataStructs

import "encoding/json"

type DBTrack struct {
	Hashval        string  `json:"hashval"`
	DateVal        string  `json:"date"`
	UTSVal         int     `json:"uts-time"`
	ArtistVal      string  `json:"artist"`
	ArtistValImage []Image `json:"artist-image"`
	TrackVal       string  `json:"track"`
	TrackValImage  []Image `json:"track-image"`
	AlbumVal       string  `json:"album"`
	ESTTime        string  `json:"EST-time"`
}

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
	Mbid       string  `json:"mtableStatus := Status{Table: *tableDescription.Table.TableName, RecordCount: *tableDescription.Table.ItemCount}bid"`
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
	//get the JSON byte data and unmarshal it into our struct.
	//Look at the structs starting at RecentTracks to understand the structure we have
	var tracks RecentTracks
	err := json.Unmarshal(*inputJson, &tracks)
	if err != nil {
		panic(err)
	}
	return &tracks
}
