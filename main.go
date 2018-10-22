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
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/bson/objectid"
	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/mongodb/mongo-go-driver/mongo/findopt"
)

// *** MONGO *** //

// ObjectID used in MongoDB
type ObjectID [12]byte

// Counter struct
type Counter struct {
	ID      objectid.ObjectID `bson:"_id"`
	Counter int               `bson:"counter"`
}

func mongoConnect() *mongo.Client {
	// Connect to MongoDB
	conn, err := mongo.Connect(context.Background(), "mongodb://localhost:27017", nil)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	return conn
}

// Check if the track already exists in the database
func urlInMongo(url string, trackColl *mongo.Collection) bool {

	// Read the documents where the trackurl field is equal to url parameter
	cursor, err := trackColl.Find(context.Background(),
		bson.NewDocument(bson.EC.String("url", url)))
	if err != nil {
		log.Fatal(err)
	}

	// 'Close' the cursor
	defer cursor.Close(context.Background())

	track := Track{}

	// Point the cursor at whatever is found
	for cursor.Next(context.Background()) {
		err = cursor.Decode(&track)
		if err != nil {
			log.Fatal(err)
		}
	}

	if track.Url == "" { // If there is an empty field, in this case, `url`, it means the track is not on the database
		return false
	}
	return true
}

// Get trackName from URL
func trackNameFromURL(url string, trackColl *mongo.Collection) string {
	// Get the trackName
	cursor, err := trackColl.Find(context.Background(),
		bson.NewDocument(bson.EC.String("trackurl", url)))

	if err != nil {
		log.Fatal(err)
	}

	defer cursor.Close(context.Background())

	dbResult := Track{}

	for cursor.Next(context.Background()) {
		err = cursor.Decode(&dbResult)
		if err != nil {
			log.Fatal(err)
		}
	}

	// TODO : i changed from dbResult.TrackName to this... edit this accordingly
	return dbResult.Url
}

// Get track counter from DB
func getTrackCounter(db *mongo.Database) int {
	counter := db.Collection("counter") // `counter` Collection

	cursor, err := counter.Find(context.Background(), nil)

	if err != nil {
		log.Fatal(err)
	}

	defer cursor.Close(context.Background())

	resCounter := Counter{}

	for cursor.Next(context.Background()) {
		err := cursor.Decode(&resCounter)
		if err != nil {
			log.Fatal(err)
		}
	}
	return resCounter.Counter
}

// Increase the track counter
func increaseTrackCounter(cnt int32, db *mongo.Database) {
	collection := db.Collection("counter") // `counter` Collection

	// This is the way to update the counter field in the document
	// Which is storen in the counter collection
	_, err := collection.UpdateOne(context.Background(), nil,
		bson.NewDocument(
			bson.EC.SubDocumentFromElements("$set",
				bson.EC.Int32("counter", cnt+1), // Increase the counter by one
			),
		),
	)
	if err != nil {
		log.Fatal(err)
	}
}

// Get all tracks
func getAllTracks(client *mongo.Client, points bool) []Track {
	db := client.Database("paragliding")  // `paragliding` Database
	collection := db.Collection("tracks") // `track` Collection

	var cursor mongo.Cursor
	var err error
	// If points boolean is true
	// Get the points for the track also
	// Otherwise don't
	if points {
		cursor, err = collection.Find(context.Background(), nil)
	} else {
		projection := findopt.Projection(bson.NewDocument(
			bson.EC.Int32("trackpoints", 0),
			bson.EC.Int32("_id", 0),
		))

		cursor, err = collection.Find(context.Background(), nil, projection)
	}

	if err != nil {
		log.Fatal(err)
	}

	defer cursor.Close(context.Background())

	resTracks := []Track{}
	resTrack := Track{}

	for cursor.Next(context.Background()) {
		err := cursor.Decode(&resTrack)
		if err != nil {
			log.Fatal(err)
		}
		resTracks = append(resTracks, resTrack) // Append each resTrack to resTracks slice
	}

	return resTracks
}

// Get track
func getTrack(client *mongo.Client, url string) Track {
	db := client.Database("igcFiles")     // `paragliding` Database
	collection := db.Collection("tracks") // `track` Collection

	cursor, err := collection.Find(context.Background(), bson.NewDocument(bson.EC.String("url", url)))

	if err != nil {
		log.Fatal(err)
	}

	resTrack := Track{}

	for cursor.Next(context.Background()) {
		err := cursor.Decode(&resTrack)
		if err != nil {
			log.Fatal(err)
		}
	}

	return resTrack

}

// Delete all tracks
func deleteAllTracks(client *mongo.Client) {
	db := client.Database("paragliding")  // `paragliding` Database
	collection := db.Collection("tracks") // `track` Collection

	// Delete the tracks
	collection.DeleteMany(context.Background(), bson.NewDocument())

	// Reset the track counter
	increaseTrackCounter(int32(0), db)

}

///////////////////////////////////////////////////////////////////////////////////////////

var timeStarted = time.Now()

var urlMap = make(map[int]string)
var mapID int
var initialID int
var uniqueId int

//IgcFiles is a slice for storing igc files
var igcFiles []Track

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
	ID        string    `json:"id"`
	Url       string    `json:"url"`
	Timestamp string    `json:"timestamp"`
	IgcTrack  igc.Track `json:"igc_track"`
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
func handler1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(r.URL.Path, "/")
	//var empty = regexp.MustCompile(``)
	var api = regexp.MustCompile(`api`)

	//Handling for /igcinfo/api
	if len(parts) != 3 || !api.MatchString(parts[2]) {
		http.Error(w, "400 - Bad Request, too many url arguments.", http.StatusBadRequest)
		return
	}
	fmt.Fprintln(w, "{"+"\"uptime\": \""+timeSince(timeStarted)+"\","+"\"info\": \"Service for IGC tracks.\","+"\"version\": \"v1\""+"}")

}

//Handling for /igcinfo/api/igc
func handler2(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	//Handling GET /igcinfo/api/igc for returning all ids storing in a slice
	case http.MethodGet:
		ids := make([]string, 0)

		client := mongoConnect()

		collection := client.Database("igcFiles").Collection("tracks")

		cursor, err := collection.Find(context.Background(), nil, nil)
		if err != nil {
			log.Fatal(err)
		}
		// 'Close' the cursor
		defer cursor.Close(context.Background())
		track := Track{}

		// Point the cursor at whatever is found
		for cursor.Next(context.Background()) {
			err = cursor.Decode(&track)
			if err != nil {
				log.Fatal(err)
			}
			ids = append(ids, track.ID)
		}

		json.NewEncoder(w).Encode(ids)

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

			track, _ := igc.ParseLocation(URL.URL)

			//uniqueId = ID
			//urlMap[uniqueId] = URL.URL
			igcFile := Track{}
			igcFile.ID = strconv.Itoa(ID)
			igcFile.IgcTrack = track
			igcFile.Url = URL.URL

			igcFile.Timestamp = time.Now().String()

			client := mongoConnect()

			collection := client.Database("igcFiles").Collection("tracks")

			// Checking for duplicates so that the user doesn't add into the database igc files with the same URL
			duplicate := urlInMongo(igcFile.Url, collection)

			if !duplicate {

				res, err := collection.InsertOne(context.Background(), igcFile)
				if err != nil {
					log.Fatal(err)
				}

				id := res.InsertedID

				if id == nil {
					http.Error(w, "", 300)
				}

				// Encoding the ID of the track that was just added to DB
				json.NewEncoder(w).Encode(igcFile.ID)

			} else {

				trackInDB := getTrack(client, igcFile.Url)
				// If there is another file in igcFilesDB with that URL return and tell the user that that IGC FILE is already in the database
				http.Error(w, "409 Conflict - The Igc File you entered is already in our database!", http.StatusConflict)
				fmt.Fprintln(w, "\nThe file you entered has the following ID: ", trackInDB.ID)
				return

			}

		}
	default:
		http.Error(w, "Not implemented", http.StatusNotImplemented)
		return
	}

}
func handler3(w http.ResponseWriter, r *http.Request) {
	//Handling /igcinfo/api/igc/<id>

	w.Header().Set("Content-Type", "application/json")
	idURL := mux.Vars(r)

	rNum, _ := regexp.Compile(`[0-9]+`)
	if !rNum.MatchString(idURL["id"]) {
		http.Error(w, "400 - Bad Request", http.StatusBadRequest)
		return
	}

	attributes := &Attributes{}

	for i := range igcFiles {

		if igcFiles[i].ID == idURL["id"] {
			attributes.HeaderDate = igcFiles[i].IgcTrack.Header.Date.String()
			attributes.Pilot = igcFiles[i].IgcTrack.Pilot
			attributes.Glider = igcFiles[i].IgcTrack.GliderType
			attributes.GliderID = igcFiles[i].IgcTrack.GliderID
			attributes.Length = trackLength(igcFiles[i].IgcTrack)

			json.NewEncoder(w).Encode(attributes)
		}
		//Handling if user type different id from ids stored

	}
	http.Error(w, "404 - The trackInfo with that id doesn't exists in IGC Files", http.StatusNotFound)

}

func handler4(w http.ResponseWriter, r *http.Request) {

	//Handling for GET /api/igc/<id>/<field>
	w.Header().Set("Content-Type", "application/json")

	urlFields := mux.Vars(r)

	var rNum, _ = regexp.Compile(`[a-zA-Z_]+`)

	if !rNum.MatchString(urlFields["field"]) {
		http.Error(w, "400 - Bad Request, wrong parameters", http.StatusBadRequest)
		return
	}

	for i := range igcFiles {

		if igcFiles[i].ID == urlFields["id"] {
			switch {
			case urlFields["field"] == "pilot":
				json.NewEncoder(w).Encode(igcFiles[i].IgcTrack.Pilot)
			case urlFields["field"] == "glider":
				json.NewEncoder(w).Encode(igcFiles[i].IgcTrack.GliderType)

			case urlFields["field"] == "glider_id":
				json.NewEncoder(w).Encode(igcFiles[i].IgcTrack.GliderID)

			case urlFields["field"] == "track_length":
				json.NewEncoder(w).Encode(trackLength(igcFiles[i].IgcTrack))

			case urlFields["field"] == "h_date":
				json.NewEncoder(w).Encode(igcFiles[i].IgcTrack.Header.Date.String())

			default:
				http.Error(w, "400 - Bad Request, the field you entered is not on our database!", http.StatusBadRequest)
				return
			}

		} else {
			http.Error(w, "400 - Bad Request, the field you entered is not on our database!", http.StatusBadRequest)
			return
		}

	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/paragliding", handler)
	r.HandleFunc("/paragliding/api", handler1)
	r.HandleFunc("/paragliding/api/track", handler2)
	r.HandleFunc("/paragliding/api/track/{id}", handler3)
	r.HandleFunc("/paragliding/api/track/{id}/{field}", handler4)
	//r.HandleFunc("/paragliding/api/ticker/latest", getApiTickerLatest)
	//r.HandleFunc("/paragliding/api/ticker/", getA)
	//r.HandleFunc("/paragliding/api/ticker/{timestamp}", handler2)

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
