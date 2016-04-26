package minitel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/pborman/uuid"
)

type Type int

const (
	App Type = iota
	User
)

type Payload struct {
	Title string `json:"title"`
	Body  string `json:"body"`

	Target struct {
		Type Type   `json:"type"`
		Id   string `json:"id"`
	}

	Action struct {
		Label string `json:"label"`
		URL   string `json:"url"`
	}
}

var (
	errNoId      = errors.New("minitel: Missing Target.Id in Payload")
	errIdNotUUID = errors.New("minitel: Target.Id not a UUID")
)

func (p Payload) Validate() error {
	if p.Target.Id == "" {
		return errNoId
	} else if res := uuid.Parse(p.Target.Id); res == nil {
		return errIdNotUUID
	}
	return nil
}

type Result struct {
	Id string `json:"id"`
}

type Client interface {
	Notify(Payload) (Result, error)
}

func New(URL string) (Client, error) {
	// Validate the URL parses.
	if _, err := url.Parse(URL); err != nil {
		return nil, err
	}
	return &client{url: URL}, nil
}

type client struct {
	url string
}

func (c *client) Notify(p Payload) (result Result, err error) {
	// Validate the payload
	if err := p.Validate(); err != nil {
		return result, err
	}

	// Prepare the buffer
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(p); err != nil {
		return result, err
	}

	// Do the HTTP POST
	resp, err := http.Post(c.url, "application/json", buf)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	// Validate http status code
	if resp.StatusCode != http.StatusCreated {
		return result, fmt.Errorf("minitel: Expected 201: Got %d", resp.StatusCode)
	}

	// Decode the response
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&result); err != nil {
		return result, err
	}
	return result, nil
}
