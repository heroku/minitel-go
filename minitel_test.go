package minitel

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
		dec.Decode(&payload)

		if payload.Validate() != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		enc := json.NewEncoder(w)
		enc.Encode(Result{Id: "727d27f8-589f-45b1-914e-dd613feaf4dc"})
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
}
