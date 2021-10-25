package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type song struct {
	Id     string `json:"id" bson:"_id,omitempty"`
	Artist string `json:"artist" bson:"artist"`
	Title  string `json:"title" bson:"title"`
	Genre  string `json:"genre" bson:"genre"`
}

/*
Legacy version where songs were all loaded into memory
var songs = []song{
	{"0", "Drake", "In My Feelings", "HipHop"},
	{"1", "Maroon 5", "Girls Like You", "Pop"},
	{"2", "Cardi B", "I Like It", "HipHop"},
	{"3", "6ix9ine", "FEFE", "Pop"},
	{"4", "Post Malone", "Better Now", "Rap"},
	{"5", "Eminem", "Lucky You", "Rap"},
	{"6", "Juice WRLD", "Lucid Dreams", "Rap"},
	{"7", "Eminem", "The Ringer", "Rap"},
	{"8", "Travis Scott", "Sicko Mode", "HipHop"},
	{"9", "Tyga", "Taste", "HipHop"},
	{"10", "Khalid & Normani", "Love Lies", "HipHop"},
	{"11", "5 Seconds Of Summer", "Youngblood", "Pop"},
	{"12", "Ella Mai", "Boo'd Up", "HipHop"},
	{"13", "Ariana Grande", "God Is A Woman", "Pop"},
	{"14", "Imagine Dragons", "Natural", "Rock"},
	{"15", "Ed Sheeran", "Perfect", "Pop"},
	{"16", "Taylor Swift", "Delicate", "Pop"},
	{"17", "Florida Georgia Line", "Simple", "Country"},
	{"18", "Luke Bryan", "Sunrise, Sunburn, Sunset", "Country"},
	{"19", "Jason Aldean", "Drowns The Whiskey", "Country"},
	{"20", "Childish Gambino", "Feels Like Summer", "HipHop"},
	{"21", "Weezer", "Africa", "Rock"},
	{"22", "Panic! At The Disco", "High Hopes", "Rock"},
	{"23", "Eric Church", "Desperate Man", "Country"},
	{"24", "Nicki Minaj", "Barbie Dreams", "Rap"},
}
*/

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("WARNING: Couldn't load .env file")
	}

	port := EnvOrInt("PORT", 80)
	mongoConnString := EnvOrString("MONGO_CONNSTRING", "")
	if mongoConnString == "" {
		log.Fatal("You must provide MONGO_CONNSTRING.")
	}
	mongoDatabase := EnvOrString("MONGO_DATABASE", "db")
	mongoCollection := EnvOrString("MONGO_COLLECTION", "col")

	log.Printf("PORT: %v", port)
	log.Println("MONGO_CONNSTRING: *SET*")
	log.Printf("MONGO_DATABASE: %v", mongoDatabase)
	log.Printf("MONGO_COLLECTION: %v", mongoCollection)

	// Attempt to initialize Cosmos connection
	log.Println("Attempting to initialize Cosmos connection...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoConnString))
	if err != nil {
		log.Fatalf("Unable to initialize Cosmos connection - %v", err)
	}
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
	log.Println("Successfully initialized Cosmos connection.")

	// Attempt to connect to a Cosmos instance
	log.Println("Attempting to connect to Cosmos...")
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Unable to connect to Cosmos - %v", err)
	}
	log.Printf("Successfully connected to Cosmos.")
	collection := client.Database(mongoDatabase).Collection(mongoCollection)

	router := gin.Default()
	// Sets the db in the gin context for every request
	router.Use(ApiMiddleware(collection))
	router.GET("/songs", getSongs)
	router.GET("/songs/:id", getSong)
	router.POST("/songs", postSong)

	router.Run(fmt.Sprintf(":%d", port))
}

func EnvOrString(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		val = def
		log.Println("WARNING: Missing " + key + " environment variable. Defaulting to: " + def)
	}
	return val
}

func EnvOrInt(key string, def int) int {
	val, err := strconv.Atoi(os.Getenv(key))
	if err != nil {
		val = def
		log.Printf("WARNING: Missing %s environment variable. Defaulting to: %d", key, def)
	}
	return val
}

func ApiMiddleware(dbColl *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("databaseCollection", dbColl)
		c.Next()
	}
}

// getSongs responds with the list of all songs as JSON.
func getSongs(c *gin.Context) {
	collection, ok := c.MustGet("databaseCollection").(*mongo.Collection)
	if !ok {
		log.Println("ERROR: Couldn't fetch database collection")
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"message": "Internal server error"})
		return
	}

	// Get all songs from the database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, bson.D{})
	if err == mongo.ErrNoDocuments {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "No songs found"})
		log.Println("WARNING: No song with id were found")
		return
	} else if err != nil {
		log.Printf("ERROR: Failure retrieving songs - %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var songs []song
	for cursor.Next(ctx) {
		//Create a song into which the single document can be decoded
		var newSong song
		err := cursor.Decode(&newSong)
		if err != nil {
			log.Fatal(err)
		}

		songs = append(songs, newSong)
	}

	c.IndentedJSON(http.StatusOK, songs)
}

// getSong locates the song whose ID value matches the id
// parameter sent by the client, then returns that song as a response.
func getSong(c *gin.Context) {
	collection, ok := c.MustGet("databaseCollection").(*mongo.Collection)
	if !ok {
		log.Println("ERROR: Couldn't fetch database collection")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Get id from query
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		log.Printf("ERROR: a valid ID was not provided - %v", err)
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "A valid ID was not provided."})
		return
	}

	// Get the song from the database
	filter := bson.M{"_id": id}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var val song
	err = collection.FindOne(ctx, filter).Decode(&val)
	if err == mongo.ErrNoDocuments {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "No song with that id was found"})
		log.Printf("WARNING: No song with id %v was found.", id)
		return
	} else if err != nil {
		log.Printf("ERROR: Failure retrieving song - %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	c.IndentedJSON(http.StatusOK, val)
}

// postSong adds a song from JSON received in the request body.
func postSong(c *gin.Context) {
	collection, ok := c.MustGet("databaseCollection").(*mongo.Collection)
	if !ok {
		log.Println("ERROR: Couldn't fetch database collection")
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	// Get the song from the request body
	var newSong song
	// Call BindJSON to bind the received JSON to newSong
	if err := c.BindJSON(&newSong); err != nil {
		return
	}

	// Insert into the database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	result, err := collection.InsertOne(ctx, newSong)
	if err != nil {
		log.Printf("ERROR: Failed to add song %v", err)
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	newSong.Id = result.InsertedID.(primitive.ObjectID).Hex()

	// Add the new song to the slice.
	songs = append(songs, newSong)
	c.IndentedJSON(http.StatusCreated, newSong)
}
