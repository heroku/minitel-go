package minitel

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPayloadCreation(t *testing.T) {
	// We only validate Id for now. If in the future
	// we need to validate other fields, we'll add them in
	// this table test.
	cases := []struct {
		Id    string
		Error error
	}{
		{"", errNoId},
		{"abc", errIdNotUUID},
		{"84838298-989d-4409-b148-6abef06df43f", nil},
	}

	for _, test := range cases {
		payload := Payload{
			Title: "Your DB is on fire!",
			Body:  "...",
		}

		payload.Target.Id = test.Id
		payload.Target.Type = App

		payload.Action.Label = "View Invoice"
		payload.Action.URL = "https://view.your.invoice/yolo"

		err := payload.Validate()

		if test.Error != err {
			t.Fatalf("Expected err == %v got %v (%+v)", test.Error, err, test)
		}
	}
}

func TestClient(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/producer/messages", func(w http.ResponseWriter, r *http.Request) {
		var payload Payload
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&payload)
		if err != nil {
			t.Errorf("ERROR: in decode: %q", err)
		}

		if payload.Validate() != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		enc := json.NewEncoder(w)
		enc.Encode(Result{Id: "727d27f8-589f-45b1-914e-dd613feaf4dc"})
	})

	mux.HandleFunc("/producer/messages/727d27f8-589f-45b1-914e-dd613feaf4dc/followups", func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		dec := json.NewDecoder(r.Body)
		err := dec.Decode(&payload)
		if err != nil {
			t.Errorf("ERROR: in decode: %q", err)
		}

		if _, ok := payload["body"]; !ok {
			t.Errorf("expected `body` parameter was not found")
		}

		w.WriteHeader(http.StatusCreated)
		enc := json.NewEncoder(w)
		enc.Encode(Result{Id: "ffff27f8-589f-45b1-914e-dd613feaf4dc"})
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	c, err := New(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	p := Payload{
		Title: "Hello",
		Body:  "DB on fire!",
	}
	p.Target.Id = "93f90f07-bbe3-433d-806d-2d01bc5ae1f2"
	res, err := c.Notify(p)
	if err != nil {
		t.Fatal(err)
	}
	if res.Id != "727d27f8-589f-45b1-914e-dd613feaf4dc" {
		t.Fatal("Expected result id to be 727d27f8-589f-45b1-914e-dd613feaf4dc (%+v)", res)
	}

	res, err = c.Followup("727d27f8-589f-45b1-914e-dd613feaf4dc", "This is a followup")
	if err != nil {
		t.Fatal(err)
	}
	if res.Id != "ffff27f8-589f-45b1-914e-dd613feaf4dc" {
		t.Fatal("Expected result id to be ffff27f8-589f-45b1-914e-dd613feaf4dc (%+v)", res)
	}
}

func TestScrubCredentials(t *testing.T) {
	u := "http://foo:bar@localhost"
	c, err := New(u)
	if err != nil {
		t.Fatalf("Error in New(%q): %q", u, err)
	}

	ac, ok := c.(*client)
	if !ok {
		t.Fatalf("Client should actually be a *client")
	}

	if strings.Contains(ac.url, "foo:bar") {
		t.Errorf("username and password were not scrubbed from URL")
	}
	if ac.user != "foo" || ac.pass != "bar" {
		t.Errorf("basic auth was not extracted from URL")
	}
}

func TestScrubCredentialsFromError(t *testing.T) {
	oldHTTP := HTTPClient
	defer func() {
		HTTPClient = oldHTTP
	}()

	// Aggressively timeout the client to give us an error containing a
	// URL
	HTTPClient = &http.Client{Timeout: time.Duration(1)}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
	}))
	defer srv.Close()

	u := strings.Replace(srv.URL, "http://", "http://foo:bar@", 1)
	c, err := New(u)
	if err != nil {
		t.Fatalf("Error in New(%q): %q", u, err)
	}

	p := Payload{
		Title: "Hello",
		Body:  "DB on fire!",
	}
	p.Target.Id = "93f90f07-bbe3-433d-806d-2d01bc5ae1f2"

	_, err = c.Notify(p)

	es := err.Error()

	// If the error contains a URL, it better not contain the creds.
	if strings.Contains(es, "http://") && strings.Contains(es, "foo:bar") {
		t.Errorf("Credentials have not been scrubbed from URL")
	}
}
