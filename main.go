package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"example.com/links-parser/parser"
)

var links = []string{
	"https://www.veriff.com/",
	"https://go.dev/",
}

type LinksWithTitle struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

type Request struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Params         url.Values
	SearchText     string
}

type ErrorObject struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type ResultSearch struct {
	Host  string           `json:"host"`
	Links []LinksWithTitle `json:"links"`
}

type ResponseObject struct {
	Error *ErrorObject   `json:"error"`
	Body  []ResultSearch `json:"body"`
}

func (r Request) errorResponse(status int, message string) {
	r.ResponseWriter.WriteHeader(status)
	responseMsg := ResponseObject{Error: &ErrorObject{Message: message, Code: status}}
	jsonContent, _ := json.Marshal(responseMsg)
	fmt.Fprintf(r.ResponseWriter, string(jsonContent))
}

func (r Request) successResponse(links []ResultSearch) {
	r.ResponseWriter.WriteHeader(http.StatusOK)
	responseMsg := ResponseObject{Body: links}
	jsonContent, _ := json.Marshal(responseMsg)
	fmt.Fprintf(r.ResponseWriter, string(jsonContent))
}

func getLinkWithTitle(requestLink string, hostLink string, searchText string, linkChan chan<- LinksWithTitle, errChan chan<- error) {

	requestURL := &url.URL{Host: hostLink[8 : len(hostLink)-1], Scheme: "https", Path: requestLink}

	response, err := http.Get(requestURL.String())
	if err != nil {
		errChan <- err
		return
	}
	defer response.Body.Close()
	bytes, _ := io.ReadAll(response.Body)
	if strings.Contains(string(bytes), searchText) {
		linkChan <- LinksWithTitle{Title: parser.ExtractTitle(bytes), Link: requestLink}
	} else {
		linkChan <- LinksWithTitle{}
	}
}

func requestAndSearch(searchText string, hostLink string, clearLinks []string) ([]LinksWithTitle, error) {
	errorChan := make(chan error)
	linkChan := make(chan LinksWithTitle)

	for _, clearLink := range clearLinks {
		go getLinkWithTitle(clearLink, hostLink, searchText, linkChan, errorChan)
	}

	var result []LinksWithTitle
	doneJobs := 0

	for {
		select {
		case err := <-errorChan:
			return nil, err
		case link := <-linkChan:
			if len(link.Link) > 0 {
				result = append(result, link)
			}
			doneJobs++
			if doneJobs == len(clearLinks) {
				return result, nil
			}
		case <-time.After(time.Second * 60):
			return nil, errors.New("Time limit exceed")
		}
	}
}

type SearchResultLinksByHost struct {
	Link  string
	Links []LinksWithTitle
}

func SearchTextByHost(link string, searchText string, result chan<- SearchResultLinksByHost, errorChan chan<- error) {
	response, err := http.Get(link)
	if err != nil {
		errorChan <- err
		return
	}
	defer response.Body.Close()
	bytes, _ := io.ReadAll(response.Body)

	clearLinks := parser.ExtractLinks(bytes)

	haveSearchText, err := requestAndSearch(searchText, link, clearLinks)
	if err != nil {
		errorChan <- err
		return
	}

	result <- SearchResultLinksByHost{Link: link, Links: haveSearchText}
}

func SearchHandler(request Request) {
	if request.SearchText == "" {
		request.errorResponse(http.StatusBadRequest, "Nothing to find")
		return
	}
	resultLinks := make(chan SearchResultLinksByHost)
	errorChan := make(chan error)

	var pageLinks []ResultSearch

	for _, link := range links {
		go SearchTextByHost(link, request.SearchText, resultLinks, errorChan)
	}

	done := 0
	for {
		select {
		case err := <-errorChan:
			request.errorResponse(http.StatusBadGateway, fmt.Sprint(err))
			return
		case result := <-resultLinks:
			pageLinks = append(pageLinks, ResultSearch{Host: result.Link, Links: result.Links})
			done++
			if done == len(links) {
				request.successResponse(pageLinks)
				return
			}
		case <-time.After(time.Second * 60):
			request.errorResponse(http.StatusRequestTimeout, fmt.Sprint("Time limit exceed"))
			return
		}
	}
}

func mainHandler(fn func(Request), method string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		requestObject := Request{
			ResponseWriter: w,
			Request:        r,
			Params:         r.URL.Query(),
			SearchText:     r.URL.Query().Get("text"),
		}

		if method != r.Method {
			requestObject.errorResponse(http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		fn(requestObject)
	}
}

func main() {
	http.HandleFunc("/search", mainHandler(SearchHandler, "GET"))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
