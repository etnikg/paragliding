package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http" //"html/template"
	"regexp"
	"strconv"
	"strings"
	"time" //"path/filepath"

	"github.com/gorilla/mux"
	igc "github.com/marni/goigc"
)

///////////////////////////////////////////////////////////////////////////////////////////

var timeStarted = time.Now()

var urlMap = make(map[int]string)
var mapID int
var initialID int
var uniqueId int

//IgcFiles is a slice for storing igc files
var igcFiles []Track

type tracks struct {
	UniqueID     string
	Pilot        string
	Glider       string
	GliderID     string
	TrackLength  float64
	Hdate        string
	Url          string
	TimeRecorded time.Time
}

func findIndex(x map[int]string, y int) bool {
	for k, _ := range x {
		if k == y {
			return false
		}
	}
	return true
}

//this function the key of the string if the map contains it, or -1 if the map does not contain the string
func searchMap(x map[int]string, y string) int {

	for k, v := range x {
		if v == y {
			return k
		}
	}
	return -1
}

func FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 4, 64)
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

//Track structure: This is structure for storing the track and to access their's id
type Track struct {
	ID  string `json:"id"`
	Url string `json:"url"`
	//	Timestamp string    `json:"timestamp"`
	IgcTrack     igc.Track `json:"igc_track"`
	TimeRecorded time.Time `json:"time_recorded"`
}

//Attributes : the info about each igc file via id
type Attributes struct {
	HeaderDate string  `json:"h_date"`
	Pilot      string  `json:"pilot"`
	Glider     string  `json:"glider"`
	GliderID   string  `json:"glider_id"`
	Length     float64 `json:"track_length"`
	TrackUrl   string  `json:"track_src_url"`
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

//Handling based in parsing url
func handler(w http.ResponseWriter, r *http.Request) {

	//Handling for /igcinfo and for /<rubbish>

	http.Error(w, "404 - Page not found!", http.StatusNotFound)

}
func handlerApi(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(r.URL.Path, "/")
	//var empty = regexp.MustCompile(``)
	var api = regexp.MustCompile(`api`)

	//Handling for /paragliding/api
	if len(parts) != 3 || !api.MatchString(parts[2]) {
		http.Error(w, "400 - Bad Request, too many url arguments.", http.StatusBadRequest)
		return
	}
	fmt.Fprintln(w, "{"+"\"uptime\": \""+timeSince(timeStarted)+"\","+"\"info\": \"Service for IGC tracks.\","+"\"version\": \"v1\""+"}")

}

//Handling for /paragliding/api/track
func handlerTrack(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	//Handling GET /paragliding/api/track for returning all ids storing in database
	case http.MethodGet:

		client := mongoConnect()

		//collection := client.Database("igcFiles").Collection("tracks")

		ids := getTrackID(client)

		fmt.Fprint(w, ids)

		// cursor, err := collection.Find(context.Background(), nil, nil)
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// // 'Close' the cursor
		// defer cursor.Close(context.Background())
		// track := Track{}

		// // Point the cursor at whatever is found
		// for cursor.Next(context.Background()) {
		// 	err = cursor.Decode(&track)
		// 	if err != nil {
		// 		log.Fatal(err)
		// 	}
		// 	ids = append(ids, track.ID)
		// }

		// json.NewEncoder(w).Encode(ids)

	case http.MethodPost:

		//handling post /igcinfo/api/igc for sending a url and returning an id for that url
		pattern := ".*.igc"

		URL := &_url{}

		var error = json.NewDecoder(r.Body).Decode(URL)
		if error != nil {
			fmt.Fprintln(w, "Error!! ", error)
			return
		}
		//rand.Seed(time.Now().UnixNano())

		res, err := regexp.MatchString(pattern, URL.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if res {
			//mapID = searchMap(urlMap, URL.URL)
			ID := rand.Intn(1000)

			track, err := igc.ParseLocation(URL.URL)
			if err != nil {
				fmt.Fprintln(w, "Error made: ", err)
				return
			}
			//uniqueId = ID
			//urlMap[uniqueId] = URL.URL
			//igcFile := Track{}
			track.UniqueID = strconv.Itoa(ID)
			//	igcFile.IgcTrack = track
			// igcFile.Url = URL.URL
			trackFile := tracks{}

			timestamp := time.Now().Second()
			timestamp = timestamp * 1000
			//	igcFile.TimeRecorded = time.Now()

			client := mongoConnect()

			collection := client.Database("igcFiles").Collection("tracks")

			// Checking for duplicates so that the user doesn't add into the database igc files with the same URL
			duplicate := urlInMongo(URL.URL, collection)

			if !duplicate {

				trackFile = tracks{
					track.UniqueID,
					track.Pilot,
					track.GliderType,
					track.GliderID,
					trackLength(track),
					track.Date.String(),
					URL.URL, time.Now()}

				res, err := collection.InsertOne(context.Background(), trackFile)
				if err != nil {
					log.Fatal(err)
				}

				id := res.InsertedID

				if id == nil {
					http.Error(w, "", 300)
				}

				// Encoding the ID of the track that was just added to DB
				fmt.Fprint(w, "{\n\"id\":\""+track.UniqueID+"\"\n}")

			} else {

				trackInDB := getTrack(client, URL.URL)
				// If there is another file in igcFilesDB with that URL return and tell the user that that IGC FILE is already in the database
				http.Error(w, "409 Conflict - The Igc File you entered is already in our database!", http.StatusConflict)
				fmt.Fprintln(w, "\nThe file you entered has the following ID: ", trackInDB.UniqueID)
				return

			}

		}
	default:
		http.Error(w, "Not implemented", http.StatusNotImplemented)
		return
	}

}
func handlerId(w http.ResponseWriter, r *http.Request) {
	//Handling /igcinfo/api/igc/<id>

	w.Header().Set("Content-Type", "application/json")
	idURL := mux.Vars(r)

	rNum, _ := regexp.Compile(`[0-9]+`)
	if !rNum.MatchString(idURL["id"]) {
		http.Error(w, "400 - Bad Request", http.StatusBadRequest)
		return
	}

	attributes := &Attributes{}

	client := mongoConnect()

	collection := client.Database("igcFiles").Collection("tracks")

	cursor, err := collection.Find(context.Background(), nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	// 'Close' the cursor
	defer cursor.Close(context.Background())

	track := &tracks{}
	//URL := &_url{}

	for cursor.Next(context.Background()) {
		err = cursor.Decode(&track)
		if err != nil {
			log.Fatal(err)
		}

		if track.UniqueID == idURL["id"] {
			attributes.HeaderDate = track.Hdate
			attributes.Pilot = track.Pilot
			attributes.Glider = track.Glider
			attributes.GliderID = track.GliderID
			attributes.Length = track.TrackLength
			attributes.TrackUrl = track.Url

			json.NewEncoder(w).Encode(attributes)
			return
		}
		//Handling if user type different id from ids stored

	}
	http.Error(w, "404 - The trackInfo with that id doesn't exists in IGC Files", http.StatusNotFound)

}

func handlerField(w http.ResponseWriter, r *http.Request) {

	//Handling for GET /api/igc/<id>/<field>
	w.Header().Set("Content-Type", "application/json")

	urlFields := mux.Vars(r)

	var rNum, _ = regexp.Compile(`[a-zA-Z_]+`)

	if !rNum.MatchString(urlFields["field"]) {
		http.Error(w, "400 - Bad Request, wrong parameters", http.StatusBadRequest)
		return
	}
	client := mongoConnect()

	collection := client.Database("igcFiles").Collection("tracks")

	cursor, err := collection.Find(context.Background(), nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	defer cursor.Close(context.Background())

	// track := Track{}
	track := &tracks{}
	for cursor.Next(context.Background()) {
		err = cursor.Decode(&track)
		if err != nil {
			log.Fatal(err)
		}

		if track.UniqueID == urlFields["id"] {

			// Mapping the track info into a Map
			fields := map[string]string{
				"pilot":         track.Pilot,
				"glider":        track.Glider,
				"glider_id":     track.GliderID,
				"track_length":  FloatToString(track.TrackLength), // Calculate the track field for the specific track and convertin it to String
				"h_date":        track.Hdate,
				"track_src_url": track.Url,
			}

			// Taking the field variable from the URL path and converting it to lower case to skip some potential errors
			field := urlFields["field"]
			field = strings.ToLower(field)

			// Searching into the map created above for the specific field that was requested
			if fieldData, ok := fields[field]; ok {
				// Encoding the data contained in the specific field saved in the map
				json.NewEncoder(w).Encode(fieldData)
				return
			} else {
				// If there is not a field like the one entered by the user. the user gets this error:
				http.Error(w, "400 - Bad Request, the field you entered is not on our database!", http.StatusBadRequest)
				return
			}
			// switch {
			// case urlFields["field"] == "pilot":
			// 	json.NewEncoder(w).Encode(track.Pilot)

			// case urlFields["field"] == "glider":
			// 	json.NewEncoder(w).Encode(track.Glider)

			// case urlFields["field"] == "glider_id":
			// 	json.NewEncoder(w).Encode(track.GliderID)

			// case urlFields["field"] == "track_length":
			// 	json.NewEncoder(w).Encode(track.TrackLength)

			// case urlFields["field"] == "h_date":
			// 	json.NewEncoder(w).Encode(track.Hdate)

			// case urlFields["field"] == "track_src_url":
			// 	json.NewEncoder(w).Encode(track.Url)

			// default:
			// 	http.Error(w, "400 - Bad Request, the field you entered is not on our database!", http.StatusBadRequest)

			// }

		} else {
			http.Error(w, "400 - Bad Request, the field you entered is not on our database!", http.StatusBadRequest)
			return
		}

	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/paragliding", handler)
	r.HandleFunc("/paragliding/api", handlerApi)
	r.HandleFunc("/paragliding/api/track", handlerTrack)
	r.HandleFunc("/paragliding/api/track/{id}", handlerId)
	r.HandleFunc("/paragliding/api/track/{id}/{field}", handlerField)
	r.HandleFunc("/paragliding/api/ticker/latest", getApiTickerLatest)
	r.HandleFunc("/paragliding/api/ticker/", getApiTicker)
	r.HandleFunc("/paragliding/api/ticker/{timestamp}", getApiTickerTimestamp)

	//fmt.Println("listening...")
	/*
		err := http.ListenAndServe(":"+os.Getenv("PORT"), r)
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}*/

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
