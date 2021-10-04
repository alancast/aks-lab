package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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

func retrieveSong(w http.ResponseWriter, r *http.Request) {
	// call "song" entity service
	songUrl := fmt.Sprint("http://localhost:9000/?id=", r.URL.Query().Get("id"))
	log.Printf("fetching song from entity service (%v)...\n", songUrl)
	songResp, err := http.Get(songUrl)
	if err != nil {
		http.Error(w, "failed to contact song service.", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// write the output
	var song song
	err = json.NewDecoder(songResp.Body).Decode(&song)
	if err != nil {
		http.Error(w, "failed to get song from entity service.", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	log.Println("successfully retrieved song.")

	// call "contracts" entity service
	contractUrl := fmt.Sprint("http://localhost:9200/?artist=", url.QueryEscape(song.Artist))
	log.Printf("fetching contract from entity service (%v)...\n", contractUrl)
	contractResp, err := http.Get(contractUrl)
	if err != nil {
		http.Error(w, "failed to contact entity service.", http.StatusInternalServerError)
		log.Println(err)
		return
	}

	// extract the payment
	var contract contract
	err = json.NewDecoder(contractResp.Body).Decode(&contract)
	if err != nil {
		http.Error(w, "failed to get contract from entity service.", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	log.Println("successfully retrieved contract.")
	song.Payment = contract.Payment

	// write the output
	bytes, err := json.Marshal(song)
	if err != nil {
		http.Error(w, "the song could not be marshalled.", http.StatusInternalServerError)
		log.Println(err)
		return
	}
	_, err = w.Write(bytes)
	if err != nil {
		http.Error(w, "the song could not be written.", http.StatusInternalServerError)
		log.Println(err)
		return
	}
}

func storeSong(w http.ResponseWriter, r *http.Request) {
	// call "song" entity service
	log.Println("federating store-song request to entity service...")
	resp, err := http.Post("http://localhost:9000", "application/json", r.Body)
	if err != nil {
		http.Error(w, "failed to contact song service.", http.StatusInternalServerError)
		return
	}

	// write the output
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "failed to get song from song service.", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(body)
	if err != nil {
		http.Error(w, "the song could not be written.", http.StatusInternalServerError)
		return
	}
}

func main() {
	http.HandleFunc("/song", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			retrieveSong(w, r)
		case "POST":
			storeSong(w, r)
		default:
			http.Error(w, "the method is not implemented.", http.StatusNotImplemented)
		}
	})
	log.Printf("listening on port %v...\n", 9100)
	err := http.ListenAndServe(":9100", nil)
	log.Fatal(err)
}
