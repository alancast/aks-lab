package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type song struct {
	Id     int    `json:"id"`
	Artist string `json:"artist"`
	Title  string `json:"title"`
	Genre  string `json:"genre"`
}

var songs = []song{
	{0, "Drake", "In My Feelings", "HipHop"},
	{1, "Maroon 5", "Girls Like You", "Pop"},
	{2, "Cardi B", "I Like It", "HipHop"},
	{3, "6ix9ine", "FEFE", "Pop"},
	{4, "Post Malone", "Better Now", "Rap"},
	{5, "Eminem", "Lucky You", "Rap"},
	{6, "Juice WRLD", "Lucid Dreams", "Rap"},
	{7, "Eminem", "The Ringer", "Rap"},
	{8, "Travis Scott", "Sicko Mode", "HipHop"},
	{9, "Tyga", "Taste", "HipHop"},
	{10, "Khalid & Normani", "Love Lies", "HipHop"},
	{11, "5 Seconds Of Summer", "Youngblood", "Pop"},
	{12, "Ella Mai", "Boo'd Up", "HipHop"},
	{13, "Ariana Grande", "God Is A Woman", "Pop"},
	{14, "Imagine Dragons", "Natural", "Rock"},
	{15, "Ed Sheeran", "Perfect", "Pop"},
	{16, "Taylor Swift", "Delicate", "Pop"},
	{17, "Florida Georgia Line", "Simple", "Country"},
	{18, "Luke Bryan", "Sunrise, Sunburn, Sunset", "Country"},
	{19, "Jason Aldean", "Drowns The Whiskey", "Country"},
	{20, "Childish Gambino", "Feels Like Summer", "HipHop"},
	{21, "Weezer", "Africa", "Rock"},
	{22, "Panic! At The Disco", "High Hopes", "Rock"},
	{23, "Eric Church", "Desperate Man", "Country"},
	{24, "Nicki Minaj", "Barbie Dreams", "Rap"},
}

func main() {
	// load .env file
	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	songs_port, err := strconv.Atoi(os.Getenv("SONGS_PORT"))
	if err != nil {
		log.Fatalf("Missing SONGS_PORT variable")
	}
	fmt.Println("SONGS_PORT:", songs_port)

	router := gin.Default()
	router.GET("/songs", getSongs)
	router.GET("/songs/:id", getSong)
	router.POST("/songs", postSong)

	router.Run(fmt.Sprintf("localhost:%d", songs_port))
}

// getSongs responds with the list of all songs as JSON.
func getSongs(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, songs)
}

// getSong locates the song whose ID value matches the id
// parameter sent by the client, then returns that song as a response.
func getSong(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Song id must be an integer"})
	}

	// Loop over the list of albums, looking for
	// an album whose ID value matches the parameter.
	for _, song := range songs {
		if song.Id == id {
			c.IndentedJSON(http.StatusOK, song)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Song not found"})
}

// postSong adds a song from JSON received in the request body.
func postSong(c *gin.Context) {
	var newSong song

	// Call BindJSON to bind the received JSON to newSong
	if err := c.BindJSON(&newSong); err != nil {
		return
	}

	// Give the new song an Id
	newSong.Id = len(songs)

	// Add the new song to the slice.
	songs = append(songs, newSong)
	c.IndentedJSON(http.StatusCreated, newSong)
}
