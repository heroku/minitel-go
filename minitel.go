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

// Type of Notification.
type Type string

// Known Telex notification target ypes.
const (
	App       Type = "app"
	User      Type = "user"
	Email     Type = "email"
	Dashboard Type = "dashboard"
)

// Notification message accepted by Telex.
type Notification struct {
	Title string `json:"title"`
	Body  string `json:"body"`

	Target Target `json:"target"`
	Action Action `json:"action"`
}

// Target portion of a Telex payload. Defined separately to ease construction
// via composite literals.
type Target struct {
	Type Type   `json:"type"`
	ID   string `json:"id"`
}

// Action portion of a Telex payload. Defined separately to ease construction
// via composite literals.
type Action struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

var (
	errNoID                 = errors.New("minitel: Missing Target.ID in Notification")
	errIDNotUUID            = errors.New("minitel: Target.ID not a UUID")
	errNoTypeSpecified      = errors.New("minitel: Missing Target.Type in Notification")
	errUnknownTypeSpecified = errors.New("minitel: Target.Type specified ")
)

// Validate that a Notification contains everything it needs to.
func (n Notification) Validate() error {
	if n.Target.ID == "" {
		return errNoID
	}
	if res := uuid.Parse(n.Target.ID); res == nil {
		return errIDNotUUID
	}
	if n.Target.Type == "" {
		return errNoTypeSpecified
	}
	switch n.Target.Type {
	case App, User, Email, Dashboard:
	default:
		return fmt.Errorf("minitel: Specified Target.Type is unknown: %s", n.Target.Type)
	}
	return nil
}

// Result from telex containing the ID of the created notification.
type Result struct {
	ID string `json:"id"`
}

// Client for communicating with telex.
type Client struct {
	url, user, pass string
	*http.Client
}

// New Telex client targeted at the telex service located at the provided URL.
func New(URL string) (*Client, error) {
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

	return &Client{
		url:    u.String(),
		user:   user,
		pass:   pass,
		Client: http.DefaultClient,
	}, nil
}

// Notify Telex.
func (c *Client) Notify(n Notification) (result Result, err error) {
	// Validate the notification before trying to send.
	if err := n.Validate(); err != nil {
		return result, err
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(n); err != nil {
		return result, err
	}

	req, err := c.postRequest(c.url+"/producer/messages", &buf)
	if err != nil {
		return result, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return result, fmt.Errorf("minitel: Expected 201: Got %d", resp.StatusCode)
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&result); err != nil {
		return result, err
	}
	return result, nil
}

// Followup adds some additional text to the previously created notification
// identified by id.
func (c *Client) Followup(id, text string) (result Result, err error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(map[string]string{"body": text}); err != nil {
		return result, err
	}

	req, err := c.postRequest(c.url+"/producer/messages/"+id+"/followups", &buf)
	if err != nil {
		return result, err
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return result, fmt.Errorf("minitel: Expected 201: Got %d", resp.StatusCode)
	}

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&result); err != nil {
		return result, err
	}
	return result, nil
}

func (c *Client) postRequest(url string, buf io.Reader) (*http.Request, error) {
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
