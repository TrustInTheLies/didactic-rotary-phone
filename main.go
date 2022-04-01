package main

import (
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type User struct {
	name string
}

var clientId string = "6827cba289b046ed823ed40ef537a468"
var clientSecret string = "524e6d924d5548a999ce68acbe92a99d"
var redirect string = "http://localhost:8080/profile"
var access map[string]interface{}
var songs []Song

type Song struct {
	Title  string
	Artist string
}

// TODO: play with channels to make a request to get your own profile after receiving a token
// like a .then() chain - https://www.newline.co/courses/build-a-spotify-connected-app/implementing-the-authorization-code-flow
func main() {
	fs := http.FileServer(http.Dir("static/"))
	http.Handle("/static/", http.StripPrefix("static", fs))
	http.HandleFunc("/login", login)
	http.HandleFunc("/profile", sendCode)
	http.HandleFunc("/refresh", refreshToken)
	http.HandleFunc("/request-page", requestInfo)
	http.HandleFunc("/send-list", getList)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err, "launching server error")
	}
}

func outputFile() {
	file, err := os.Create("songs.csv")
	if err != nil {
		log.Fatal(err, "creating file error")
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	fmt.Println(len(songs))
	fmt.Println(songs)
	for _, song := range songs {
		str := []string{"Title: " + song.Title, "Artist: " + song.Artist}
		err := w.Write(str)
		if err != nil {
			log.Fatal(err, "writing to CSV error")
		}
	}

}

func getList(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		fmt.Println()
		err := json.NewDecoder(r.Body).Decode(&songs)
		if err != nil {
			log.Fatal(err, "decoding answer error")
		}
		outputFile()
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Headers", "*")

}

func requestInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(fmt.Sprintf("%v", access["access_token"]))
}

func refreshToken(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	token := fmt.Sprintf("%v", access["refresh_token"])
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", token)
	req, err := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err, "refresh error")
	}
	auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(clientId+":"+clientSecret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", auth)
	req.Header.Set("Access-Control-Allow-Origin", "*")
	do, sendErr := client.Do(req)
	if sendErr != nil {
		log.Fatal(sendErr)
	}
	decodErr := json.NewDecoder(do.Body).Decode(&access)
	if decodErr != nil {
		log.Fatal(decodErr, "decode error")
	}
}

func sendCode(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	params, err := url.Parse(r.URL.String())
	if err != nil {
		log.Fatal(err, "url parsing error")
	}

	code := strings.Join(params.Query()["code"], "")
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirect)

	req, reqErr := http.NewRequest("POST", "https://accounts.spotify.com/api/token", strings.NewReader(data.Encode()))
	if reqErr != nil {
		log.Fatal(reqErr, "post code error")
	}
	// TODO: may explode
	auth := "Basic " + base64.StdEncoding.EncodeToString([]byte(clientId+":"+clientSecret))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", auth)
	do, sendErr := client.Do(req)
	if sendErr != nil {
		log.Fatal(sendErr)
	}
	decodErr := json.NewDecoder(do.Body).Decode(&access)
	if decodErr != nil {
		log.Fatal(decodErr, "decode error")
	}
	http.Redirect(w, r, "/retrieve-songs", 301)
}

func login(w http.ResponseWriter, r *http.Request) {
	req, err := http.NewRequest("GET", "https://accounts.spotify.com/authorize", nil)
	if err != nil {
		log.Fatal(err, "get error")
	}
	q := req.URL.Query()
	q.Add("client_id", clientId)
	q.Add("response_type", "code")
	q.Add("redirect_uri", redirect)
	q.Add("scope", "user-read-private user-library-read playlist-read-collaborative playlist-read-private")
	q.Add("show_dialog", "true")
	req.URL.RawQuery = q.Encode()
	http.Redirect(w, r, req.URL.String(), 301)
}
