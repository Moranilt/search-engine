package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	DB *sqlx.DB
}

type Request struct {
	ResponseWriter http.ResponseWriter
	Request        *http.Request
	Params         url.Values
	SearchPhrase   string
}

type ErrorObject struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

type ResultSearch struct {
	Host  string           `json:"host"`
	Links []LinksWithTitle `json:"links"`
}

type HostWithEndpoints struct {
	Host         string                   `json:"host"`
	IsSearchable bool                     `json:"is_searchable"`
	Endpoints    []EndpointBySearchPhrase `json:"endpoints"`
}

type ResultBodyType interface {
	[]ResultSearch | []HostWithEndpoints | []string | int
}

type ResponseObject[T ResultBodyType] struct {
	Error *ErrorObject `json:"error"`
	Body  T            `json:"body"`
}

func (r Request) errorResponse(status int, message string) {
	r.ResponseWriter.WriteHeader(status)
	responseMsg := ResponseObject[[]string]{Error: &ErrorObject{Message: message, Code: status}}
	jsonContent, _ := json.Marshal(responseMsg)
	fmt.Fprintf(r.ResponseWriter, string(jsonContent))
}

func (r Request) successResponse(links []ResultSearch) {
	r.ResponseWriter.WriteHeader(http.StatusOK)
	responseMsg := ResponseObject[[]ResultSearch]{Body: links}
	jsonContent, _ := json.Marshal(responseMsg)
	fmt.Fprintf(r.ResponseWriter, string(jsonContent))
}

func (r Request) successResponseHostsPost(amountOfAdded int) {
	r.ResponseWriter.WriteHeader(http.StatusOK)
	responseMsg := ResponseObject[int]{Body: amountOfAdded}
	jsonContent, _ := json.Marshal(responseMsg)
	fmt.Fprintf(r.ResponseWriter, string(jsonContent))
}

func (r Request) successResponseHostsGet(hosts []HostWithEndpoints) {
	r.ResponseWriter.WriteHeader(http.StatusOK)
	responseMsg := ResponseObject[[]HostWithEndpoints]{Body: hosts}
	jsonContent, _ := json.Marshal(responseMsg)
	fmt.Fprintf(r.ResponseWriter, string(jsonContent))
}

func mainHandler(fn func(Request), method string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")

		request := Request{
			ResponseWriter: w,
			Request:        r,
			Params:         r.URL.Query(),
			SearchPhrase:   r.URL.Query().Get("text"),
		}

		if method != r.Method {
			request.errorResponse(http.StatusMethodNotAllowed, MethodNotAllowed)
			return
		}

		fn(request)
	}
}
