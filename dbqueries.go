package main

import (
	"github.com/jmoiron/sqlx"
)

type Host struct {
	Id           int    `db:"id"`
	Name         string `db:"name"`
	IsSearchable bool   `db:"is_searchable"`
	CreatedAt    string `db:"created_at"`
	tx           *sqlx.Tx
}

func (h Host) GetEndpoints(db *sqlx.DB) (endpoints []EndpointBySearchPhrase) {
	db.Select(&endpoints, "SELECT name as path, titles.value as title FROM endpoints INNER JOIN titles ON titles.endpoint_id=endpoints.id WHERE endpoints.host_id=$1", h.Id)
	return
}

func (h Host) GetEndpointsWithoutSearchPhrase(db *sqlx.DB, searchPhrase string) (endpoints []string) {
	db.Select(&endpoints, `SELECT 
	e.name as path
	FROM endpoints e 
	INNER JOIN hosts h ON h.id=e.host_id 
	INNER JOIN titles ON titles.endpoint_id=e.id
	LEFT JOIN (
		SELECT 
		ep.endpoint_id, 
		phrases.name as phrase 
		FROM phrases 
		INNER JOIN endpoints_phrases ep 
		ON ep.phrase_id=phrases.id
	) as p ON e.id=p.endpoint_id WHERE p.phrase != $1 OR p.phrase IS NULL`, searchPhrase)

	return
}

func (h Host) GetEndpointsWithSearchPhrase(db *sqlx.DB, searchPhrase string) (endpoints []EndpointBySearchPhrase) {
	db.Select(&endpoints, `SELECT e.name as path, titles.value as title FROM endpoints e 
	INNER JOIN hosts h ON h.id=e.host_id 
	INNER JOIN endpoints_phrases ep ON ep.endpoint_id=e.id 
	INNER JOIN phrases ON ep.phrase_id=phrases.id 
	INNER JOIN titles ON titles.endpoint_id=e.id
	WHERE phrases.name=$1 AND h.name=$2`, searchPhrase, h.Name)
	return
}

func (h Host) EndpointContains(db *sqlx.DB, endpoint string) (exists bool) {
	db.Get(&exists, "SELECT EXISTS (SELECT FROM endpoints WHERE host_id=$1 AND name=$2)", h.Id, endpoint)
	return
}

func (h *Host) MustBegin(db *sqlx.DB) {
	h.tx = db.MustBegin()
}

func (h Host) StoreEndpointByPhrase(endpoint string, searchPhrase string, title string) {
	h.tx.MustExec(
		"SELECT create_endpoint_phrase_title($1, $2, $3, $4)",
		h.Id,
		endpoint,
		searchPhrase,
		title,
	)
}

func (h Host) NewEndpoint(endpoint string, title string) {
	h.tx.MustExec(
		"SELECT create_endpoint_title($1,$2,$3)",
		h.Id,
		endpoint,
		title,
	)
}

func (h *Host) Commit() error {
	err := h.tx.Commit()
	h.tx = nil
	return err
}

type EndpointBySearchPhrase struct {
	Path  string `db:"path" json:"path"`
	Title string `db:"title" json:"title"`
}

func (e EndpointBySearchPhrase) GetSearchPhrases(db *sqlx.DB) []string {
	var phrases []string
	db.Select(&phrases, `
	SELECT phrases.name FROM phrases 
	INNER JOIN endpoints_phrases ep ON ep.phrase_id=phrases.id
	INNER JOIN endpoints ON ep.endpoint_id=endpoints.id
	WHERE endpoints.name=$1
	`, e.Path)

	return phrases
}

const (
	CreateHostQuery                  = "INSERT INTO hosts (name, is_searchable) VALUES ($1, false) ON CONFLICT (name) DO NOTHING"
	SelectNameOfHosts                = "SELECT name FROM hosts"
	SelectAllFromHosts               = "SELECT * FROM hosts"
	SelectHostByName                 = "SELECT * FROM hosts WHERE name=$1"
	SelectAllEndpointsBySearchPhrase = `
	SELECT e.name as path, titles.value as title FROM endpoints e 
	INNER JOIN hosts h ON h.id=e.host_id 
	INNER JOIN endpoints_phrases ep ON ep.endpoint_id=e.id 
	INNER JOIN phrases ON ep.phrase_id=phrases.id 
	INNER JOIN titles ON titles.endpoint_id=e.id 
	WHERE phrases.name=$1 AND h.name=$2
	`
	ChangeHostsIsSearchableState  = "UPDATE hosts SET is_searchable=true WHERE name=$1"
	GetEndpointsWithSearchPhrases = `SELECT endpoints.name as name, array_agg(phrases.name)as phrases FROM endpoints 
	INNER JOIN endpoints_phrases ep ON endpoints.id=ep.endpoint_id 
	INNER JOIN phrases ON phrases.id=ep.phrase_id GROUP BY endpoints.name`
)
