package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/jmoiron/sqlx"
)

const (
	GetMethod     = "GET"
	PostMethod    = "POST"
	PutMethod     = "PUT"
	PatchMethod   = "PATCH"
	DeleteMethod  = "DELETE"
	OptionsMethod = "OPTIONS"
	HeadMethod    = "HEAD"
)

type Route struct {
	Path    string
	Handler func(Request)
}

type Routes struct {
	existingRoutes map[string]bool
	routes         map[string][]Route
}

func (r *Routes) storeRoute(method string, route string, handler func(Request)) {
	r.existingRoutes[route] = true
	r.routes[method] = append(r.routes[method], Route{Path: route, Handler: handler})
}

func (r Routes) Exists(route string) bool {
	return r.existingRoutes[route]
}

func (r Routes) GetRoutes(method string) []Route {
	return r.routes[method]
}

type Repository struct {
	DB     *sqlx.DB
	Routes *Routes
}

func NewRepository(db *sqlx.DB) *Repository {
	routes := Routes{existingRoutes: make(map[string]bool), routes: make(map[string][]Route)}
	return &Repository{Routes: &routes, DB: db}
}

func (repository Repository) GetRoutes(method string) []Route {
	return repository.Routes.GetRoutes(method)
}

func (repository *Repository) storeRoute(method string, route string, handler func(Request)) {
	repository.Routes.storeRoute(method, route, handler)
}

func (repository Repository) Get(route string, handler func(Request)) {
	repository.storeRoute(GetMethod, route, handler)
}

func (repository Repository) Post(route string, handler func(Request)) {
	repository.storeRoute(PostMethod, route, handler)
}

func (repository Repository) Put(route string, handler func(Request)) {
	repository.storeRoute(PutMethod, route, handler)
}

func (repository Repository) Patch(route string, handler func(Request)) {
	repository.storeRoute(PatchMethod, route, handler)
}

func (repository Repository) Delete(route string, handler func(Request)) {
	repository.storeRoute(DeleteMethod, route, handler)
}

func (repository Repository) Options(route string, handler func(Request)) {
	repository.storeRoute(OptionsMethod, route, handler)
}

func (repository Repository) Head(route string, handler func(Request)) {
	repository.storeRoute(HeadMethod, route, handler)
}

func (repository *Repository) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	request := Request{
		ResponseWriter: w,
		Request:        r,
		Params:         r.URL.Query(),
		SearchPhrase:   r.URL.Query().Get("text"),
	}

	routesByMethod := repository.GetRoutes(r.Method)

	for _, route := range routesByMethod {

		if r.URL.Path == route.Path {
			route.Handler(request)
			return
		}
	}

	if repository.Routes.Exists(r.URL.Path) {
		request.errorResponse(http.StatusMethodNotAllowed, MethodNotAllowed)
		return
	}
	request.errorResponse(http.StatusNotFound, "Page not found")
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
