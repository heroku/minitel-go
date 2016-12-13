package minitel

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/pborman/uuid"
)

type Type int

const (
	App Type = iota
	User
	Email
)

var HTTPClient = http.DefaultClient

func (t Type) MarshalJSON() ([]byte, error) {
	switch t {
	case App:
		return []byte(`"app"`), nil
	case User:
		return []byte(`"user"`), nil
	case Email:
		return []byte(`"email"`), nil
	default:
		return []byte(""), fmt.Errorf("unknown Type: %d", t)
	}
}

func (t *Type) UnmarshalJSON(raw []byte) error {
	switch string(raw) {
	case `"app"`:
		*t = App
		return nil
	case `"user"`:
		*t = User
		return nil
	case `"email"`:
		*t = Email
		return nil
	default:
		return errors.New("can't unmarshal Type")
	}
}

type Payload struct {
	Title string `json:"title"`
	Body  string `json:"body"`

	Target struct {
		Type Type   `json:"type"`
		Id   string `json:"id"`
	} `json:"target"`

	Action struct {
		Label string `json:"label"`
		URL   string `json:"url"`
	} `json:"action"`
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
	Notify(p Payload) (Result, error)
	Followup(id string, body string) (Result, error)
}

func New(URL string) (Client, error) {
	// Validate the URL parses.
	u, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	var user, pass string
	if u.User != nil {
		user = u.User.Username()
		pass, _ = u.User.Password()
		u.User = nil
	}

	return &client{
		url:  u.String(),
		user: user,
		pass: pass,
	}, nil
}

type client struct {
	url, user, pass string
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

	req, err := c.postRequest(c.url+"/producer/messages", buf)
	if err != nil {
		return result, err
	}

	// Do the HTTP POST
	resp, err := HTTPClient.Do(req)

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

func (c *client) Followup(id, body string) (result Result, err error) {
	// Prepare the buffer
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(map[string]string{"body": body}); err != nil {
		return result, err
	}

	req, err := c.postRequest(c.url+"/producer/messages/"+id+"/followups", buf)
	if err != nil {
		return result, err
	}

	// Do the HTTP Post
	resp, err := HTTPClient.Do(req)

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

func (c *client) postRequest(url string, buf io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return nil, err
	}

	if c.user != "" || c.pass != "" {
		req.SetBasicAuth(c.user, c.pass)
	}
	req.Header.Set("content-type", "application/json")
	return req, err
}
