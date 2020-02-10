package main

import (
	"container/list"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"hash/fnv"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Global vars, includes database collection, map and list for LRU, and mutex
var (
	db    *mongo.Collection
	m     map[string]*list.Element
	lru   *list.List
	mtx   sync.RWMutex
	once  sync.Once
	limit int = 512
)

// Initializes map and LRU
func init() {
	once.Do(initMapLRU)
}
func initMapLRU() {
	m = make(map[string]*list.Element)
	lru = list.New()
}

// key/value for LRU
type kv struct {
	k string
	v string
}

// the data
type miniAndLongURL struct {
	MiniURL string `json:"miniurl"`
	LongURL string `json:"long"`
	Hits    uint64 `json:"hits"`
}

// main driver for the app
// connects to the database, initializes the server, routes requests
func main() {
	e := dbConnect()
	if e != nil {
		log.Fatalf(e.Error())
	}
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type", "Content-Length"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowAllOrigins:  false,
		AllowOriginFunc:  func(origin string) bool { return true },
	}))

	server.Use(static.Serve("/", static.LocalFile("./client/build", true)))
	server.Use(static.Serve("/stats", static.LocalFile("./client/build", true)))

	server.POST("/api/new", buildURL)
	server.GET("/:m", redirURL)
	server.POST("/api/current", getStats)

	err := server.Run(getPort())
	if err != nil {
		panic(err)
	}
}

// defines the port to be used based on env variable, uses 8080 if no port specified
func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return ":" + port
}

// HTTP Request functions
// buildURL creates the MiniURL, adds to database, and returns the mini to the user
func buildURL(context *gin.Context) {
	url, e := convertHTTPBodyToMiniStruct(context.Request.Body)
	if e != nil {
		context.JSON(500, e.Error())
		return
	}

	e = url.makeMini()
	updateLRU(url.MiniURL, kv{k: url.MiniURL, v: url.LongURL})
	url.MiniURL = context.Request.Host + "/" + url.MiniURL
	if e != nil && e.Error() == "That mini already exists" {
		context.JSON(200, struct {
			MiniURL string `json:"miniurl"`
			Error   string `json:"error"`
		}{
			MiniURL: url.MiniURL,
			Error:   e.Error(),
		})
		return
	} else if e != nil {
		context.JSON(500, e.Error())
		return
	}
	context.JSON(200, url)
}

// redirURL is used for redirecting the user who clicks on the MiniURL
// it also increments the number of hits
func redirURL(context *gin.Context) {
	mini := context.Params.ByName("m")
	if found := mapFind(mini).Value.(kv); found.v != "" {
		context.Redirect(303, found.v)
		updateLRU(mini, found)
		updateMini(bson.M{"miniurl": bson.M{"$eq": mini}}, bson.M{"$inc": bson.M{"hits": 1}})
		return
	}
	if url, ok := findMini(bson.M{"miniurl": bson.M{"$eq": mini}}); ok == nil {
		context.Redirect(303, url.LongURL)
		updateLRU(mini, kv{k: mini, v: url.LongURL})
		updateMini(bson.M{"miniurl": bson.M{"$eq": url.MiniURL}}, bson.M{"$inc": bson.M{"hits": 1}})
		return
	}
	context.Status(404)
}

// getStats returns the miniAndLongURL struct that contains:
// - the mini url
// - the original url
// - the number of hits the mini has
func getStats(context *gin.Context) {
	url, e := convertHTTPBodyToMiniStruct(context.Request.Body)
	if e != nil {
		context.JSON(500, e.Error())
		return
	}
	if len(url.MiniURL) > 6 {
		url.MiniURL = url.MiniURL[len(url.MiniURL)-7 : len(url.MiniURL)-1]
	}
	if dburl, ok := findMini(bson.M{"miniurl": bson.M{"$eq": url.MiniURL}}); ok == nil {
		context.JSON(200, dburl)
		updateLRU(dburl.MiniURL, kv{k: dburl.MiniURL, v: dburl.LongURL})
		return
	}
	context.Status(404)
}

// These functions handle the creation of the mini URL
// makeMini takes in a miniAndLongURL struct that only needs the LongURL field initialized
// it calls helper functions to finish creation of the struct and add it to the database
// it returns an error if adding the struct to the database encounters an error
func (u *miniAndLongURL) makeMini() error {
	u.MiniURL = encode(hash32(u.LongURL))
	if found := mapFind(u.MiniURL).Value.(kv).v; found != "" {
		updateLRU(u.MiniURL, kv{k: u.MiniURL, v: u.LongURL})
		return errors.New("That mini already exists")
	}
	if found, _ := findMini(bson.M{"miniurl": bson.M{"$eq": u.MiniURL}}); found.MiniURL != u.MiniURL {
		u.Hits = 0
		err := addMini(u)
		updateLRU(u.MiniURL, kv{k: u.MiniURL, v: u.LongURL})
		return err
	}
	return errors.New("That mini already exists")
}

// hash32 uses a 32 bit FNV function to generate the miniURL
// the hash is returned as an array of bytes
func hash32(s string) []byte {
	hash := fnv.New32a()
	hash.Write([]byte(s))

	m := hash.Sum(nil)
	return m
}

// encode transforms the result from hash32 into a URL SAFE base64 encoding
// this base64 string is returned
func encode(m []byte) string {
	ss := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(m)
	return ss
}

// Converts the body of an HTTP Request to the miniAndLongURL struct
// This returns the built struct and an error if one is encountered
func convertHTTPBodyToMiniStruct(httpBody io.ReadCloser) (miniAndLongURL, error) {
	body, e := ioutil.ReadAll(httpBody)

	if e != nil {
		return miniAndLongURL{}, e
	}
	var url miniAndLongURL
	e = json.Unmarshal(body, &url)
	if e != nil {
		return miniAndLongURL{}, e
	}
	return url, nil
}

// findMini is the helper function that looks up a miniURL in the database
// The parameter query needs to be formatted in bson.M so that mongodb can properly look for the mini
// This function returns a miniAndLongURL struct (nil if not found) and an error if one occurs
func findMini(query bson.M) (*miniAndLongURL, error) {
	var url miniAndLongURL
	e := db.FindOne(context.Background(), query).Decode(&url)
	return &url, e
}

// addMini adds a miniAndLongURL struct to the database
// returns an error if one occurs
func addMini(url *miniAndLongURL) error {
	_, e := db.InsertOne(context.Background(), *url)
	return e
}

// updates a miniAndLongURL object in the database based on the filter and update params
// returns an error if one occurs
func updateMini(filter, update bson.M) error {
	_, e := db.UpdateOne(context.Background(), filter, update)
	return e
}

// uses the enviornment variables to connect to the database
// initializes the global "db" variable (at the top)
// set env variables for MONGOURI=<the mongoURI>, DBNAME=<name of the database>, DBCOL=<name of the database collection>
func dbConnect() error {
	muri := os.Getenv("MONGOURI")
	dbname := os.Getenv("DBNAME")
	col := os.Getenv("DBCOL")
	if muri == "" {
		return errors.New("Could not obtain the Mongo URI. Ensure it is set in the ENV under the name \"MONGOURI\"\n")
	}
	if dbname == "" {
		return errors.New("Could not obtain the DB Collection. Ensure it is set in the ENV under the name \"DBCOL\"\n")
	}
	if col == "" {
		return errors.New("Could not obtain the Mongo URI. Ensure it is set in the ENV under the name \"MONGOURI\"\n")
	}
	clientOptions := options.Client().ApplyURI(muri)
	client, e := mongo.Connect(context.TODO(), clientOptions)
	if e != nil {
		return e
	}
	if e := client.Ping(context.TODO(), nil); e != nil {
		return e
	}
	db = client.Database(dbname).Collection(col)
	return nil
}

// LRU functions
// updateLRU removes oldest if limit is reached, adds new elements when used
func updateLRU(mini string, val kv) {
	if found := mapFind(mini); found.Value.(kv).v != "" {
		mtx.Lock()
		lru.MoveToFront(found)
		mtx.Unlock()
		return
	}
	tmp := list.Element{Value: val}
	mtx.Lock()
	lru.PushFront(&tmp)
	m[mini] = &tmp
	if lru.Len() > limit {
		del := lru.Back()
		if del != nil {
			delete(m, del.Value.(*list.Element).Value.(kv).k)
			lru.Remove(del)
		}
	}
	mtx.Unlock()
}

// mapFind checks for existing keys in the map
func mapFind(mini string) *list.Element {
	mtx.RLock()
	defer mtx.RUnlock()
	if val, ok := m[mini]; ok {
		return val
	}
	ret := list.Element{Value: kv{v: ""}}
	return &ret
}
