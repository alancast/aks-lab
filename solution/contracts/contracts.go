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

type contract struct {
	Artist  string  `json:"artist"`
	Payment float64 `json:"payment"`
}

var contracts = []contract{
	{"Drake", 0.2},
	{"Taylor Swift", 0.25},
	{"Khalid & Normani", 0.1},
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("WARNING: Couldn't load .env file")
	}

	contracts_port, err := strconv.Atoi(os.Getenv("CONTRACTS_PORT"))
	if err != nil {
		log.Fatalf("Missing CONTRACTS_PORT variable")
	}
	log.Println("CONTRACTS_PORT:", contracts_port)

	router := gin.Default()
	router.GET("/contracts", getContract)

	router.Run(fmt.Sprintf(":%d", contracts_port))
}

// getContract locates the contract whose ID value matches the id
// parameter sent by the client, then returns that contract as a response.
func getContract(c *gin.Context) {
	artist, ok := c.GetQuery("artist")
	if !ok {
		fmt.Println("Returning all contracts as no artist was specified")
		c.IndentedJSON(http.StatusOK, contracts)
		return
	}

	// Loop over the list of contracts, looking for
	// a contract whose artist value matches the parameter.
	for _, contract := range contracts {
		if contract.Artist == artist {
			c.IndentedJSON(http.StatusOK, contract)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Contract not found"})
}
