package main

import (
	"github.com/jmoiron/sqlx"
)

type Repository struct {
	DB *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{DB: db}
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
