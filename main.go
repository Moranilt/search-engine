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
	"https://www.spacex.com/",
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

type ResponseObject struct {
	Error *ErrorObject                `json:"error"`
	Body  map[string][]LinksWithTitle `json:"body"`
}

func (r Request) errorResponse(status int, message string) {
	r.ResponseWriter.WriteHeader(http.StatusBadGateway)
	responseMsg := ResponseObject{Error: &ErrorObject{Message: message, Code: status}}
	jsonContent, _ := json.Marshal(responseMsg)
	fmt.Fprintf(r.ResponseWriter, string(jsonContent))
}

func (r Request) successResponse(links map[string][]LinksWithTitle) {
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
	if strings.Index(string(bytes), searchText) != -1 {
		linkChan <- LinksWithTitle{Title: parser.ExtractTitle(bytes), Link: requestLink}
	} else {
		linkChan <- LinksWithTitle{}
	}
}

func requestAndSearch(request Request, hostLink string, clearLinks []string) ([]LinksWithTitle, error) {
	errorChan := make(chan error)
	linkChan := make(chan LinksWithTitle)

	uniqueLinks := make(map[string]bool)

	for _, clearLink := range clearLinks {
		if uniqueLinks[clearLink] {
			continue
		} else {
			uniqueLinks[clearLink] = true
			go getLinkWithTitle(clearLink, hostLink, request.SearchText, linkChan, errorChan)
		}
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

func SearchHandler(request Request) {
	if request.SearchText == "" {
		request.errorResponse(http.StatusBadRequest, "Nothing to find")
		return
	}
	pageLinks := make(map[string][]LinksWithTitle)
	for _, link := range links {
		response, err := http.Get(link)
		if err != nil {
			request.errorResponse(http.StatusBadGateway, fmt.Sprintf("Error: %s", err))
			return
		}
		defer response.Body.Close()
		bytes, _ := io.ReadAll(response.Body)

		clearLinks := parser.ExtractLinks(bytes)

		haveSearchText, err := requestAndSearch(request, link, clearLinks)

		if err != nil {
			request.errorResponse(http.StatusBadGateway, fmt.Sprint(err))
			return
		}

		pageLinks[link] = haveSearchText
	}

	request.successResponse(pageLinks)
}

func mainHandler(fn func(Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		requestObject := Request{
			ResponseWriter: w,
			Request:        r,
			Params:         r.URL.Query(),
			SearchText:     r.URL.Query().Get("text"),
		}
		fn(requestObject)
	}
}

func main() {

	http.HandleFunc("/search", mainHandler(SearchHandler))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
