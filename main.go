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

	"github.com/Moranilt/rou"
	"github.com/Moranilt/search-engine/parser"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const (
	BodyIsNotValid   = "Request body is not valid"
	MethodNotAllowed = "Method not allowed"
)

type LinksWithTitle struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

func getLinkWithTitleBySearch(requestLink string, hostLink string, searchPhrase string, linkChan chan<- LinksWithTitle, errChan chan<- error) {
	requestURL := &url.URL{Host: hostLink[8 : len(hostLink)-1], Scheme: "https", Path: requestLink}

	response, err := http.Get(requestURL.String())
	if err != nil {
		errChan <- err
		return
	}
	defer response.Body.Close()
	bytes, _ := io.ReadAll(response.Body)
	if strings.Contains(string(bytes), searchPhrase) {
		linkChan <- LinksWithTitle{Title: parser.ExtractTitle(bytes), Link: requestLink}
	} else {
		linkChan <- LinksWithTitle{}
	}
}

func requestAndSearch(searchPhrase string, hostLink string, clearLinks []string) ([]LinksWithTitle, error) {
	errorChan := make(chan error)
	linkChan := make(chan LinksWithTitle)

	for _, clearLink := range clearLinks {
		go getLinkWithTitleBySearch(clearLink, hostLink, searchPhrase, linkChan, errorChan)
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

func (repository Repository) searchPhraseByHost(host Host, searchPhrase string, result chan<- SearchResultLinksByHost, errorChan chan<- error) {

	var searchResult []LinksWithTitle

	eWithSearchPhrase := host.GetEndpointsWithSearchPhrase(repository.DB, searchPhrase)
	eWithoutSearchPhrase := host.GetEndpointsWithoutSearchPhrase(repository.DB, searchPhrase)

	for _, endpoint := range eWithSearchPhrase {
		searchResult = append(searchResult, LinksWithTitle{Title: endpoint.Title, Link: endpoint.Path})
	}

	if len(eWithoutSearchPhrase) > 0 {
		haveSearchPhrase, err := requestAndSearch(searchPhrase, host.Name, eWithoutSearchPhrase)
		if err != nil {
			errorChan <- err
			return
		}
		host.MustBegin(repository.DB)
		for _, endpoint := range haveSearchPhrase {
			host.StoreEndpointByPhrase(endpoint.Link, searchPhrase, endpoint.Title)
		}
		host.Commit()

		if err != nil {
			errorChan <- err
			return
		}
		searchResult = append(searchResult, haveSearchPhrase...)
	}

	result <- SearchResultLinksByHost{Link: host.Name, Links: searchResult}
}

func (repository Repository) SearchHandler(request *rou.Context) {
	searchPhrase := request.Params().Get("text")
	if searchPhrase == "" {
		request.ErrorJSONResponse(http.StatusBadRequest, "Nothing to find")
		return
	}
	resultLinks := make(chan SearchResultLinksByHost)
	errorChan := make(chan error)

	var pageLinks []ResultSearch

	var hosts []Host
	repository.DB.Select(&hosts, SelectAllFromHosts)

	for _, link := range hosts {
		go repository.searchPhraseByHost(link, searchPhrase, resultLinks, errorChan)
	}

	done := 0
	for {
		select {
		case err := <-errorChan:
			request.ErrorJSONResponse(http.StatusBadGateway, fmt.Sprint(err))
			return
		case result := <-resultLinks:
			done++
			if len(result.Links) > 0 {
				pageLinks = append(pageLinks, ResultSearch{Host: result.Link, Links: result.Links})
			}
			if done == len(hosts) {
				request.SuccessJSONResponse(pageLinks)
				return
			}
		case <-time.After(time.Second * 60):
			request.ErrorJSONResponse(http.StatusRequestTimeout, fmt.Sprint("Time limit exceed"))
			return
		}
	}
}

func (repository Repository) POST_HostsHandler(request *rou.Context) {
	requestBody, err := io.ReadAll(request.Request().Body)

	if err != nil {
		request.ErrorJSONResponse(http.StatusBadRequest, BodyIsNotValid)
		return
	}

	var hosts []string

	err = json.Unmarshal(requestBody, &hosts)

	if err != nil {
		request.ErrorJSONResponse(http.StatusBadRequest, BodyIsNotValid)
		return
	}

	var successCounter int

	for _, host := range hosts {
		result := repository.DB.MustExec(CreateHostQuery, host)
		r, _ := result.RowsAffected()
		if r > 0 {
			successCounter++
		}
	}

	request.SuccessJSONResponse(successCounter)
}

func (repository Repository) GET_HostsHandler(request *rou.Context) {
	var hosts []Host

	repository.DB.Select(&hosts, "SELECT * FROM hosts")

	var hostsWithEndpoints []HostWithEndpoints

	for _, host := range hosts {
		hostsWithEndpoints = append(
			hostsWithEndpoints,
			HostWithEndpoints{
				Host:         host.Name,
				IsSearchable: host.IsSearchable,
				Endpoints:    host.GetEndpoints(repository.DB),
			},
		)
	}

	request.SuccessJSONResponse(hostsWithEndpoints)
}

// Add all endpoints by host to DB and activate host
func (repository Repository) ActivateHosts(request *rou.Context) {
	defer request.Request().Body.Close()
	body, _ := io.ReadAll(request.Request().Body)
	var hosts []string
	err := json.Unmarshal(body, &hosts)

	if err != nil {
		request.ErrorJSONResponse(http.StatusBadRequest, fmt.Sprint(err))
		return
	}

	var selectAllHostsByName strings.Builder
	selectAllHostsByName.WriteString("SELECT * FROM hosts WHERE ")
	for i, host := range hosts {
		if i == len(hosts)-1 {
			selectAllHostsByName.WriteString(fmt.Sprintf("name='%s'", host))
		} else {
			selectAllHostsByName.WriteString(fmt.Sprintf("name='%s' OR ", host))
		}
	}

	var dbHosts []Host
	repository.DB.Select(&dbHosts, selectAllHostsByName.String())

	var addedEndpoints int

	for _, host := range dbHosts {
		response, err := http.Get(host.Name)
		if err != nil {
			request.ErrorJSONResponse(http.StatusBadGateway, fmt.Sprint(err))
			return
		}
		defer response.Body.Close()
		body, _ := io.ReadAll(response.Body)
		clearLinks := parser.ExtractLinks(body)

		host.MustBegin(repository.DB)
		resultChan := make(chan LinksWithTitle)

		for _, link := range clearLinks {
			go getLinkWithTitle(host, link, resultChan)
		}

		for i := 0; i < len(clearLinks); i++ {
			linkWithTitle := <-resultChan
			host.NewEndpoint(linkWithTitle.Link, linkWithTitle.Title)
		}
		err = host.Commit()
		if err != nil {
			request.ErrorJSONResponse(http.StatusBadGateway, fmt.Sprint(err))
			return
		}
		repository.DB.MustExec(ChangeHostsIsSearchableState, host.Name)
	}

	request.SuccessJSONResponse(addedEndpoints)
}

func getLinkWithTitle(host Host, link string, resultChan chan<- LinksWithTitle) {
	requestURL := &url.URL{Host: host.Name[8 : len(host.Name)-1], Scheme: "https", Path: link}

	response, err := http.Get(requestURL.String())
	if err != nil {
		return
	}

	defer response.Body.Close()

	html, _ := io.ReadAll(response.Body)
	if err != nil {
		return
	}
	resultChan <- LinksWithTitle{Link: link, Title: parser.ExtractTitle(html)}
}

func main() {
	db, err := sqlx.Connect("postgres", "user=root password=123456 dbname=search_engine sslmode=disable")

	if err != nil {
		log.Fatal(err)
	}

	repository := NewRepository(db)
	router := rou.NewRouter()

	router.Get("/search", repository.SearchHandler)
	router.Get("/hosts/list", repository.GET_HostsHandler)
	router.Post("/hosts/add", repository.POST_HostsHandler)
	router.Post("/hosts/activate", repository.ActivateHosts)
	log.Fatal(router.RunServer(":8080"))
}
