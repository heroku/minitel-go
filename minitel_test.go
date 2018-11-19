package minitel

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/heroku/minitel-go/miniteltest"
)

func TestNotificationCreation(t *testing.T) {
	// We only validate Id for now. If in the future
	// we need to validate other fields, we'll add them in
	// this table test.
	cases := []struct {
		ID    string
		Error error
	}{
		{"", errNoID},
		{"abc", errIDNotUUID},
		{"84838298-989d-4409-b148-6abef06df43f", nil},
	}

	for _, test := range cases {
		notification := Notification{
			Title: "Your DB is on fire!",
			Body:  "...",
		}

		notification.Target.ID = test.ID
		notification.Target.Type = App

		notification.Action.Label = "View Invoice"
		notification.Action.URL = "https://view.your.invoice/yolo"

		err := notification.Validate()

		if test.Error != err {
			t.Fatalf("Expected err == %v got %v (%+v)", test.Error, err, test)
		}
	}
}

func TestClient(t *testing.T) {
	ts := miniteltest.NewServer()
	defer ts.Close()

	notifyUUID := "727d27f8-589f-45b1-914e-dd613feaf4dc"
	ts.ExpectNotify(miniteltest.GenerateHTTPResponse(t, notifyUUID, http.StatusCreated))

	followUpUUID := "ffff27f8-589f-45b1-914e-dd613feaf4dc"
	ts.ExpectFollowup(miniteltest.GenerateHTTPResponse(t, followUpUUID, http.StatusCreated))

	c, err := New(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	n := Notification{
		Title: "Hello",
		Body:  "DB on fire!",
		Target: Target{
			Type: App,
		},
	}
	n.Target.ID = "93f90f07-bbe3-433d-806d-2d01bc5ae1f2"
	res, err := c.Notify(n)
	if err != nil {
		t.Fatal(err)
	}
	if res.ID != notifyUUID {
		t.Fatalf("Expected result id to be 727d27f8-589f-45b1-914e-dd613feaf4dc (%+v)", res)
	}

	res, err = c.Followup("727d27f8-589f-45b1-914e-dd613feaf4dc", "This is a followup")
	if err != nil {
		t.Fatal(err)
	}
	if res.ID != followUpUUID {
		t.Fatalf("Expected result id to be ffff27f8-589f-45b1-914e-dd613feaf4dc (%+v)", res)
	}

	if finished := ts.ExpectDone(time.Second); !finished {
		t.Fatalf("Expected no pending expectations, but some still exist")
	}
}

func TestScrubCredentials(t *testing.T) {
	u := "http://foo:bar@localhost"
	c, err := New(u)
	if err != nil {
		t.Fatalf("Error in New(%q): %q", u, err)
	}

	if strings.Contains(c.url, "foo:bar") {
		t.Errorf("username and password were not scrubbed from URL")
	}
	if c.user != "foo" || c.pass != "bar" {
		t.Errorf("basic auth was not extracted from URL")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name         string
		notification Notification
		wantErr      error
	}{
		{
			name:         "no target ID",
			notification: Notification{},
			wantErr:      errNoID,
		},
		{
			name:         "target ID not UUID",
			notification: Notification{Target: Target{ID: "123"}},
			wantErr:      errIDNotUUID,
		},
		{
			name:         "no target type specified",
			notification: Notification{Target: Target{ID: "bc31ed62-0204-40e5-86cf-b25a001b20db"}},
			wantErr:      errNoTypeSpecified,
		},
		{
			name:         "target type App",
			notification: Notification{Target: Target{ID: "bc31ed62-0204-40e5-86cf-b25a001b20db", Type: App}},
			wantErr:      nil,
		},
		{
			name:         "target type Email",
			notification: Notification{Target: Target{ID: "bc31ed62-0204-40e5-86cf-b25a001b20db", Type: Email}},
			wantErr:      nil,
		},
		{
			name:         "target type User",
			notification: Notification{Target: Target{ID: "bc31ed62-0204-40e5-86cf-b25a001b20db", Type: User}},
			wantErr:      nil,
		},
		{
			name:         "target type Dashboard",
			notification: Notification{Target: Target{ID: "bc31ed62-0204-40e5-86cf-b25a001b20db", Type: Dashboard}},
			wantErr:      nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if gotErr := test.notification.Validate(); test.wantErr != gotErr {
				t.Fatalf("want error: %v, got %v", test.wantErr, gotErr)
			}
		})
	}

}
