package miniteltest

import (
	"testing"
	"time"

	minitel "github.com/heroku/minitel-go"
)

var (
	n = minitel.Notification{
		Title: "Hello",
		Body:  "DB on fire!",
		Target: minitel.Target{
			Type: minitel.App,
			ID:   "93f90f07-bbe3-433d-806d-2d01bc5ae1f2",
		},
	}
)

func TestWait(t *testing.T) {
	for _, tc := range []struct {
		name     string
		prep     func(ts *TestServer)
		process  func(c *minitel.Client)
		finished bool
	}{
		{
			name:     "None",
			prep:     nil,
			finished: true,
		},
		{
			name: "One Unprocessed Notification",
			prep: func(ts *TestServer) {
				ts.ExpectNotify(nil)
				ts.ExpectNotify(nil)
			},
			process: func(c *minitel.Client) {
				c.Notify(n)
			},
			finished: false,
		},
		{
			name: "One Unprocessed Followup",
			prep: func(ts *TestServer) {
				ts.ExpectFollowup(nil)
				ts.ExpectFollowup(nil)
			},
			process: func(c *minitel.Client) {
				c.Followup("testid", "testtext")
			},
			finished: false,
		},
		{
			name: "Mixed None Left",
			prep: func(ts *TestServer) {
				ts.ExpectNotify(nil)
				ts.ExpectFollowup(nil)
			},
			process: func(c *minitel.Client) {
				c.Notify(n)
				c.Followup("testid", "testtext")
			},
			finished: true,
		},
		{
			name: "Mixed One Unprocessed Followup",
			prep: func(ts *TestServer) {
				ts.ExpectNotify(nil)
				ts.ExpectFollowup(nil)
			},
			process: func(c *minitel.Client) {
				c.Notify(n)
			},
			finished: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ts := NewServer()
			defer ts.Close()

			if tc.prep != nil {
				tc.prep(ts)
			}

			if tc.process != nil {
				c, err := minitel.New(ts.URL)
				if err != nil {
					t.Fatalf("unable to construct minitel client from TestServer.URL (%q): %s", ts.URL, err)
				}
				tc.process(c)
			}

			if finished := ts.Wait(100 * time.Millisecond); finished != tc.finished {
				t.Errorf("expected Wait() to return %t, but got %t", tc.finished, finished)
			}

			if tc.finished {
				ts.ExpectDone(t)
			}
		})
	}
}

func TestExpectNotify(t *testing.T) {
	ts := NewServer()
	defer ts.Close()

	ts.ExpectNotify(nil)

	c, err := minitel.New(ts.URL)
	if err != nil {
		t.Fatal("unable to setup test client: ", err)
	}

	r, err := c.Notify(n)
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}
	if r.ID == "" {
		t.Error("expected the ID to not be blank, but it was")
	}
}

type testExpectDoneFatal bool

func (t *testExpectDoneFatal) Fatal(...interface{}) {
	*t = true
}

func TestExpectDone(t *testing.T) {
	ts := NewServer()
	defer ts.Close()

	ts.ExpectNotify()
	var tdf testExpectDoneFatal = false
	ts.ExpectDone(&tdf)

	if !bool(tdf) {
		t.Fatalf("expected tdf to be true, got: %v", tdf)
	}
}

func TestExpectNoNotify(t *testing.T) {
	ts := NewServer()
	defer ts.Close()

	c, err := minitel.New(ts.URL)
	if err != nil {
		t.Fatal("unable to setup test client: ", err)
	}

	if _, err := c.Notify(n); err == nil {
		t.Fatal("expected error but was nil")
	}
}

func TestExpectFollowup(t *testing.T) {
	ts := NewServer()
	defer ts.Close()

	ts.ExpectFollowup(nil)

	c, err := minitel.New(ts.URL)
	if err != nil {
		t.Fatal("unable to setup test client: ", err)
	}

	r, err := c.Followup("testid", "testtext")
	if err != nil {
		t.Fatal("unexpected error: ", err)
	}
	if r.ID == "" {
		t.Error("expected the ID to not be blank, but it was")
	}
}

func TestExpectNoFollowup(t *testing.T) {
	ts := NewServer()
	defer ts.Close()

	c, err := minitel.New(ts.URL)
	if err != nil {
		t.Fatal("unable to setup test client: ", err)
	}

	if _, err := c.Followup("testid", "testtext"); err == nil {
		t.Fatal("expected error but was nil")
	}
}
