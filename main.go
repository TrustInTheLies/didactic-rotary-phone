package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var clientId string = "6827cba289b046ed823ed40ef537a468"
var clientSecret string = "524e6d924d5548a999ce68acbe92a99d"
var redirect string = "http://localhost:8080/profile"
var access map[string]interface{}

func main() {
	http.HandleFunc("/login", login)
	http.HandleFunc("/profile", sendCode)
	http.HandleFunc("/refresh", refreshToken)
	http.ListenAndServe(":8080", nil)
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
	do, sendErr := client.Do(req)
	if sendErr != nil {
		log.Fatal(sendErr)
	}
	decodErr := json.NewDecoder(do.Body).Decode(&access)
	if decodErr != nil {
		log.Fatal(decodErr, "decode error")
	}
	fmt.Fprint(w, access["access_token"])
}

func sendCode(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	params, err := url.Parse(r.URL.String())
	if err != nil {
		log.Fatal(err, "url parsing error")
	}

	code := strings.Join(params.Query()["code"], "")
	fmt.Println(code)
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
	fmt.Fprint(w, access["access_token"])
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
	// fmt.Println(r)
	//fmt.Println(req.URL.String())
	//resp, clientErr := client.Do(req)
	//if clientErr != nil {
	//	return
	//}
	//fmt.Fprint(w, resp)
}
