package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"log"
	"net/http"
)

type BbRequest struct {
	Filter int    `json:"filter"`
	Intro  int    `json:"intro"`
	Query  string `json:"query"`
}

type BbResponse struct {
	BadQuery int    `json:"bad_query"`
	Error    int    `json:"error"`
	Query    string `json:"query"`
	Text     string `json:"text"`
}

type BbClient struct {
	client *http.Client
}

func NewBbClient() *BbClient {
	bbClientTransport := http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	httpClient := http.Client{Transport: &bbClientTransport}
	return &BbClient{&httpClient}
}

func (bbClient *BbClient) Run(query string) (*BbResponse, error) {
	reqBodyBytes, err := json.Marshal(BbRequest{
		Filter: 1,
		Intro:  0,
		Query:  query,
	})
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(
		http.MethodPost,
		envPanic("API_URL"),
		bytes.NewReader(reqBodyBytes),
	)
	request.Header.Add(
		"User-Agent",
		envPanic("REQ_USERAGENT"),
	)
	request.Header.Add("Content-Type", "application/json")
	request.Host = envPanic("REQ_HOST")
	if err != nil {
		return nil, err
	}

	response, err := bbClient.client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()
	var bbResponse BbResponse
	err = json.NewDecoder(response.Body).Decode(&bbResponse)
	if err != nil {
		return nil, err
	}

	log.Printf("\"%s\" -> Status: %s\n", query, response.Status)
	return &bbResponse, nil
}
