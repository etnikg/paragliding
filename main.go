package main

import (
	"fmt"
	"log"
	"net/http" //"html/template"
	"os"
	"strconv"
	"time" //"path/filepath"

	"github.com/gorilla/mux"
	igc "github.com/marni/goigc"
)

var timeStarted = time.Now()
var uniqueID int

type tracks struct {
	UniqueID     string
	Pilot        string
	Glider       string
	GliderID     string
	TrackLength  float64
	Hdate        string
	URL          string
	TimeRecorded time.Time
}

//FloatToString : convert a float number to a string
func FloatToString(inputNum float64) string {

	return strconv.FormatFloat(inputNum, 'f', 4, 64)
}

type _url struct {
	URL string `json:"url"`
}

//Calculating the total length of track
func trackLength(track igc.Track) float64 {

	totalDistance := 0.0

	for i := 0; i < len(track.Points)-1; i++ {
		totalDistance += track.Points[i].Distance(track.Points[i+1])
	}

	return totalDistance
}

//Calculating uptime based on ISO 8601
func timeSince(t time.Time) string {

	Decisecond := 100 * time.Millisecond
	Day := 24 * time.Hour

	ts := time.Since(t)
	sign := time.Duration(1)

	ts += Decisecond / 2
	d := sign * (ts / Day)
	ts = ts % Day
	h := ts / time.Hour
	ts = ts % time.Hour
	m := ts / time.Minute
	ts = ts % time.Minute
	s := ts / time.Second
	ts = ts % time.Second
	f := ts / Decisecond
	y := d / 365
	return fmt.Sprintf("P%dY%dD%dH%dM%d.%dS", y, d, h, m, s, f)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/paragliding", handler)
	r.HandleFunc("/paragliding/api", handlerAPI)
	//Handling Track
	r.HandleFunc("/paragliding/api/track", handlerTrack)
	r.HandleFunc("/paragliding/api/track/{id}", handlerID)
	r.HandleFunc("/paragliding/api/track/{id}/{field}", handlerField)
	//Handling ticker
	r.HandleFunc("/paragliding/api/ticker/latest", getApiTickerLatest)
	r.HandleFunc("/paragliding/api/ticker", getApiTicker)
	r.HandleFunc("/paragliding/api/ticker/{timestamp}", getApiTickerTimestamp)
	//Handling the webhooks
	r.HandleFunc("/paragliding/api/webhook/new_track/", webhookNewTrack)
	r.HandleFunc("/paragliding/api/webhook/new_track/{webhook_id}", webhookID)
	//Handling the admin part
	r.HandleFunc("/paragliding/admin/api/tracks_count", adminAPITracksCount)
	r.HandleFunc("/paragliding/admin/api/tracks", adminAPITracks)
	r.HandleFunc("/pargliding/admin/api/webhooks", adminAPIWebhookTrigger)

	//fmt.Println("listening...")

	err := http.ListenAndServe(":"+os.Getenv("PORT"), r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

	/*if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}*/
}
