package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type song struct {
	Id      int     `json:"id"`
	Title   string  `json:"title"`
	Artist  string  `json:"artist"`
	Payment float64 `json:"payment"`
	Genre   string  `json:"genre"`
}

type contract struct {
	Artist  string  `json:"artist"`
	Payment float64 `json:"payment"`
}

var songServiceBaseUrl = "http://songs"
var contractServiceBaseUrl = "http://contracts"

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("WARNING: Couldn't load .env file")
	}

	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("Missing PORT variable")
	}
	log.Println("PORT:", port)

	val, present := os.LookupEnv("SONG_SERVICE_BASE_URL")
	if present {
		songServiceBaseUrl = val
	}
	log.Println("songServiceBaseUrl:", songServiceBaseUrl)

	val, present = os.LookupEnv("CONTRACT_SERVICE_BASE_URL")
	if present {
		contractServiceBaseUrl = val
	}
	log.Println("contractServiceBaseUrl:", contractServiceBaseUrl)

	router := gin.Default()
	router.GET("/health", getHealth)
	router.GET("/songs", getSong)
	router.POST("/songs", postSong)
	router.Run(fmt.Sprintf(":%d", port))
}
func getHealth(c *gin.Context) {
	c.Status(http.StatusOK)
}

// getSong locates the song whose ID value matches the id
// parameter sent by the client, then returns that song as a response.
func getSong(c *gin.Context) {
	songServiceUrl := songServiceBaseUrl
	id_str, ok := c.GetQuery("id")
	if ok {
		// Check if id is a valid int
		id, err := strconv.Atoi(id_str)
		if err != nil || id < 0 {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Song id must be an integer greater than 0"})
			return
		}

		songServiceUrl = fmt.Sprint(songServiceUrl, "/", id_str)
	}

	// Query to get the song(s)
	songsResponse, err := http.Get(songServiceUrl)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Failed to contact song service"))
		log.Println(err)
		return
	}

	// Check if response is a failure
	if songsResponse.StatusCode > 299 {
		c.AbortWithStatus(songsResponse.StatusCode)
		body, err := io.ReadAll(songsResponse.Body)
		if err == nil {
			log.Println("Song Response Error: ", string(body))
		}
		return
	}

	// Parse songs from response
	songs := []song{}
	var singlesong song
	responseBodyData, err := ioutil.ReadAll(songsResponse.Body)
	responseBodyReader1 := bytes.NewReader(responseBodyData)
	responseBodyReader2 := bytes.NewReader(responseBodyData)
	err = json.NewDecoder(responseBodyReader1).Decode(&songs)
	errSingle := json.NewDecoder(responseBodyReader2).Decode(&singlesong)
	if err != nil && errSingle != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Failed to decode songs"))
		log.Println(err)
		return
	}

	if err != nil && errSingle == nil {
		songs = append(songs, singlesong)
	}

	// Get payment info for each song
	for index := range songs {
		songs[index].Payment = getArtistPayment(songs[index].Artist)
	}

	c.IndentedJSON(http.StatusOK, songs)
}

// Queries to get the contract for the given artist and return the payment value
// If no contrat exists it returns a default payment of 0
func getArtistPayment(artist string) float64 {
	contractServiceURL := fmt.Sprint(contractServiceBaseUrl, "?artist=", url.QueryEscape(artist))

	// Query to get the contract
	contractResponse, err := http.Get(contractServiceURL)
	if err != nil {
		log.Println("ERROR: getting contract for artist: ", artist)
		log.Println("ERROR: message ", err)
		return 0
	}

	// Make sure contract query was successful
	if contractResponse.StatusCode > 299 {
		log.Println("WARNING: No contract found for artist: ", artist)
		return 0
	}

	// Json decode contract values
	var contract contract
	err = json.NewDecoder(contractResponse.Body).Decode(&contract)
	if err != nil {
		log.Println("ERROR: decoding contract for artist: ", artist)
		log.Println("ERROR: message ", err)
		return 0
	}

	return contract.Payment
}

// postSong adds a song from JSON received in the request body.
func postSong(c *gin.Context) {
	// Send song to song service
	songsResponse, err := http.Post(songServiceBaseUrl, "application/json", c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Failed to contact song service"))
		log.Println(err)
		return
	}

	// Check if response is a failure
	if songsResponse.StatusCode > 299 {
		c.AbortWithStatus(songsResponse.StatusCode)
		body, err := io.ReadAll(songsResponse.Body)
		if err == nil {
			log.Println("Error posting song: ", string(body))
		}
		return
	}

	// Decode created song and forward response
	var song song
	err = json.NewDecoder(songsResponse.Body).Decode(&song)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, errors.New("Failed to decode song from post body"))
		log.Println(err)
		return
	}

	c.IndentedJSON(songsResponse.StatusCode, song)
}
