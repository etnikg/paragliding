package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	igc "github.com/marni/goigc"
)

//Handling based in parsing url
func handler(w http.ResponseWriter, r *http.Request) {

	//Handling for /igcinfo and for /<rubbish>

	http.Error(w, "404 - Page not found!", http.StatusNotFound)

}
func handlerAPI(w http.ResponseWriter, r *http.Request) {
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

		ids := getTrackID(client)

		fmt.Fprint(w, ids)

	case http.MethodPost:

		//handling post /igcinfo/api/igc for sending a url and returning an id for that url
		pattern := ".*.igc"

		URL := &_url{}

		var error = json.NewDecoder(r.Body).Decode(URL)
		if error != nil {
			fmt.Fprintln(w, "Error!! ", error)
			return
		}
		res, err := regexp.MatchString(pattern, URL.URL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if res {
			ID := rand.Intn(1000)

			track, err := igc.ParseLocation(URL.URL)
			if err != nil {
				fmt.Fprintln(w, "Error made: ", err)
				return
			}

			track.UniqueID = strconv.Itoa(ID)

			trackFile := tracks{}

			timestamp := time.Now().Second()
			timestamp = timestamp * 1000

			client := mongoConnect()

			collection := client.Database("igcfiles").Collection("tracks")

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

				triggerWhenTrackIsAdded()

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
func handlerID(w http.ResponseWriter, r *http.Request) {
	//Handling /igcinfo/api/igc/<id>
	if r.Method != "GET" {
		http.Error(w, "501 - Method not implemented", http.StatusNotImplemented)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	idURL := mux.Vars(r)

	rNum, _ := regexp.Compile(`[0-9]+`)
	if !rNum.MatchString(idURL["id"]) {
		http.Error(w, "400 - Bad Request", http.StatusBadRequest)
		return
	}

	client := mongoConnect()

	collection := client.Database("igcfiles").Collection("tracks")

	cursor, err := collection.Find(context.Background(), nil, nil)
	if err != nil {
		log.Fatal(err)
	}

	// 'Close' the cursor
	defer cursor.Close(context.Background())

	track := tracks{}
	//URL := &_url{}

	for cursor.Next(context.Background()) {
		err = cursor.Decode(&track)
		if err != nil {
			log.Fatal(err)
		}

		if track.UniqueID == idURL["id"] {
			fmt.Fprint(w, "{\n\"H_date\":\""+track.Hdate+"\",\n\"pilot\":\""+track.Pilot+"\",\n\"glider\":\""+track.Glider+"\",\n\"glider_id\":\""+track.GliderID+"\",\n\"length\":\""+FloatToString(track.TrackLength)+"\",\n\"track_src_url\":\""+track.URL+"\"\n}")

		} else {
			//Handling if user type different id from ids stored
			http.Error(w, "404 - The trackInfo with that id doesn't exists in our database ", http.StatusNotFound)

		}

	}
	http.Error(w, "404 - The trackInfo with that id doesn't exists in IGC Files", http.StatusNotFound)

}

func handlerField(w http.ResponseWriter, r *http.Request) {

	//Handling for GET /api/igc/<id>/<field>
	w.Header().Set("Content-Type", "application/json")

	urlFields := mux.Vars(r)

	var rNum, _ = regexp.Compile(`[a-zA-Z_]+`)

	//attributes := &Attributes{}

	// Regular Expression for IDs

	regExID, _ := regexp.Compile("[0-9]+")

	if !regExID.MatchString(urlFields["id"]) {
		http.Error(w, "400 - Bad Request, you entered an invalid ID in URL.", http.StatusBadRequest)
		return
	}

	if !rNum.MatchString(urlFields["field"]) {
		http.Error(w, "400 - Bad Request, wrong parameters", http.StatusBadRequest)
		return
	}
	client := mongoConnect()

	trackDB := tracks{}

	trackDB = getTrack1(client, urlFields["id"], w)
	// Taking the field variable from the URL path and converting it to lower case to skip some potential errors
	field := urlFields["field"]

	switch field {
	case "pilot":
		fmt.Fprint(w, trackDB.Pilot)
	case "glider":
		fmt.Fprint(w, trackDB.Glider)
	case "glider_id":
		fmt.Fprint(w, trackDB.GliderID)
	case "h_date":
		fmt.Fprint(w, trackDB.Hdate)
	case "track_length":
		fmt.Fprint(w, trackDB.TrackLength)
	case "track_src_url":
		fmt.Fprint(w, trackDB.URL)
	default:
		http.Error(w, "", 404)
	}

}
