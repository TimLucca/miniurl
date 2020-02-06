package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io"
	"io/ioutil"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// For preDB testing
var db map[string]miniAndLongURL

func init() {
	db = make(map[string]miniAndLongURL)
}

// the data
type miniAndLongURL struct {
	MiniURL string `json:"mini"`
	LongURL string `json:"long"`
	Hits    uint64
}

func (u *miniAndLongURL) makeMini() {
	u.MiniURL = encode(hash32(u.LongURL))
	if _, found := db[u.MiniURL]; !found {
		u.Hits = 0
		db[u.MiniURL] = *u
	}
}

func mainCLI() {
	var s string
	fmt.Print("Enter url: ")
	fmt.Scan(&s)
	mini := encode(hash32(s))
	fmt.Printf("Minified URL: https://ex-mini.com/%s\n", mini)
}

func hash32(s string) []byte {
	hash := fnv.New32a()
	hash.Write([]byte(s))

	m := hash.Sum(nil)
	return m
}

func encode(m []byte) string {
	ss := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(m)
	return ss
}

func main() {
	server := gin.Default()
	server.Use(cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type", "Content-Length"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowAllOrigins:  false,
		AllowOriginFunc:  func(origin string) bool { return true },
	}))
	// server.LoadHTMLGlob("./client/*.html")

	// server.GET("/", func(c *gin.Context) {
	// 	c.HTML(200, "index.html", nil)
	// })

	server.POST("/api/new", buildURL)
	server.GET("/:m", redirURL)
	server.POST("/api/current", getStats)

	err := server.Run(":8080")
	if err != nil {
		panic(err)
	}
}

func buildURL(context *gin.Context) {
	url, e := convertHTTPBodyToMiniStruct(context.Request.Body)
	if e != nil {
		context.JSON(500, e.Error())
		return
	}

	url.makeMini()
	context.JSON(200, url)
}

func redirURL(context *gin.Context) {
	mini := context.Params.ByName("m")
	if url, ok := db[mini]; ok {
		url.Hits += 1
		db[mini] = url
		context.Redirect(303, url.LongURL)
		return
	}
	log.Println("Not found")
	log.Println(db)
	context.Status(404)
}

func getStats(context *gin.Context) {
	url, e := convertHTTPBodyToMiniStruct(context.Request.Body)
	if e != nil {
		context.JSON(500, e.Error())
		return
	}
	if dburl, ok := db[url.MiniURL]; ok {
		context.JSON(200, dburl)
		return
	}
	context.Status(404)
}

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
