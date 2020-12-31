package main

import (
	"encoding/xml"
	"html/template"
	"io/ioutil"
	"net/http"
	"sync"
)

var wg sync.WaitGroup

type mapIndex struct {
	LocURL []string `xml:"sitemap>loc"` // follow xml to get correct location
}

type newsInfo struct {
	Titles  []string `xml:"url>news>title"`
	PubDate []string `xml:"url>news>publication_date"` // can be replaced accordingly
	LocURL  []string `xml:"url>loc"`
}

type mMap struct {
	Keyword  string
	Location string
}

type webPage struct {
	Title    string
	newsInfo map[string]mMap
}

func newsHandler(w http.ResponseWriter, r *http.Request) {
	var m mapIndex
	response, _ := http.Get("*YOUR WEBMAP HERE*") // refer to readme to find webmap
	bytes, _ := ioutil.ReadAll(response.Body)
	xml.Unmarshal(bytes, &m)
	nMap := make(map[string]mMap)
	response.Body.Close()
	queue := make(chan newsInfo, 30)

	for _, Location := range m.LocURL {
		wg.Add(1)
		go newsRoutine(queue, Location)
	}
	wg.Wait()
	close(queue)

	for element := range queue {
		for i := range element.PubDate {
			nMap[element.Titles[i]] = mMap{element.PubDate[i], element.LocURL[i]}
		}
	}

	p := webPage{Title: "Local News Database", newsInfo: nMap}
	t, _ := template.ParseFiles("newsaggtemplate.html")
	t.Execute(w, p)
}

func newsRoutine(ch chan newsInfo, Location string) {
	defer wg.Done()
	var n newsInfo
	response, _ := http.Get(Location)
	bytes, _ := ioutil.ReadAll(response.Body)
	xml.Unmarshal(bytes, &n)
	response.Body.Close()
	ch <- n
}

func main() {
	http.HandleFunc("/", newsHandler)
	http.HandleFunc("/agg/", newsHandler)
	http.ListenAndServe(":8000", nil)
}
