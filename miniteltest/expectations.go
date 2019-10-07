package miniteltest

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

// TestServer implements the basic interactions with Minitel for the
// purposes of testing. Use ExpectNotify and ExpectFollowup to register
// http.Responses with the handler for controlled testing of responses. Or call
// those methods with nil to get a generic response. The ExpectDone() method can be
// used to ensure that all expectations have happened within the provided
// timeout, which is useful for when the client is used async.
type TestServer struct {
	*httptest.Server

	sync.Mutex
	notifyResponses   []*http.Response
	followupResponses []*http.Response
}

// Here so we don't have to import minitel
type result struct {
	ID uuid.UUID `json:"id"`
}

// NewServer returns a prepared TestServer which should be used like a httptest.Server
//
//    ts := NewServer()
//    defer ts.Close()
//
func NewServer() *TestServer {
	var ts TestServer
	ts.Server = httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				ts.Lock()
				defer ts.Unlock()
				if r.Method != http.MethodPost {
					http.Error(w, "Unexpected Method: "+r.Method, http.StatusInternalServerError)
					return
				}
				if r.URL.Path == "/producer/messages" {
					ts.notifyHandler(w, r)
					return
				}

				if strings.HasPrefix(r.URL.Path, "/producer/messages/") && strings.HasSuffix(r.URL.Path, "/followups") {
					ts.followupHandler(w, r)
					return
				}

				http.Error(w, "Unexpected path: "+r.URL.Path, http.StatusInternalServerError)
			},
		),
	)
	return &ts
}

func doResponse(resp *http.Response, w http.ResponseWriter) {
	if resp == nil {
		w.WriteHeader(http.StatusCreated)
		enc := json.NewEncoder(w)
		enc.Encode(result{ID: uuid.New()})
		return
	}
	for k, v := range resp.Header {
		for _, v1 := range v {
			w.Header().Add(k, v1)
		}
	}
	w.WriteHeader(resp.StatusCode)
	if resp.Body != nil {
		io.Copy(w, resp.Body)
	}
}

func (ts *TestServer) notifyHandler(w http.ResponseWriter, r *http.Request) {
	var n interface{}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&n)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(ts.notifyResponses) == 0 {
		http.Error(w, "No Notify Response Expecations", http.StatusInternalServerError)
		return
	}
	resp := ts.notifyResponses[0]
	ts.notifyResponses = ts.notifyResponses[1:]
	doResponse(resp, w)
}

func (ts *TestServer) followupHandler(w http.ResponseWriter, r *http.Request) {
	var p struct {
		Body string `json:"body"`
	}
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if p.Body == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(ts.followupResponses) == 0 {
		http.Error(w, "No Followup Response Expectations", http.StatusInternalServerError)
		return
	}
	resp := ts.followupResponses[0]
	ts.followupResponses = ts.followupResponses[1:]
	doResponse(resp, w)
}

// ExpectNotify and send the supplied responses. If the responses are nil an empty
// response with a random ID and a http.StatusCreated is sent
func (ts *TestServer) ExpectNotify(r ...*http.Response) {
	ts.Lock()
	defer ts.Unlock()

	if r == nil {
		ts.notifyResponses = append(ts.notifyResponses, nil)
		return
	}

	ts.notifyResponses = append(ts.notifyResponses, r...)
}

// ExpectFollowup and send the supplied responses. If the responses are nil an
// empty response with a random ID and a http.StatusCreated is sent
func (ts *TestServer) ExpectFollowup(r ...*http.Response) {
	ts.Lock()
	defer ts.Unlock()

	if r == nil {
		ts.followupResponses = append(ts.followupResponses, nil)
		return
	}

	ts.followupResponses = append(ts.followupResponses, r...)
}

// ExpectDone waits up to max duration for all notify and followup responses to
// be sent. Returns true if they have been sent. If they haven't been sent after
// the max duration then return false.
func (ts *TestServer) ExpectDone(max time.Duration) bool {
	expire := time.NewTimer(max)
	defer expire.Stop()
	for {
		ts.Lock()
		lnr := len(ts.notifyResponses)
		lfr := len(ts.followupResponses)
		ts.Unlock()
		if lnr == 0 && lfr == 0 {
			return true
		}
		time.Sleep(max / 20)
		select {
		case <-expire.C:
			return false
		default:
		}
	}
}

// GenerateHTTPResponse purposes with the given Result and StatusCode.
// This is here to reduce boilerplate construction in the common case.
func GenerateHTTPResponse(t *testing.T, id uuid.UUID, c int) *http.Response {
	rb, err := json.Marshal(result{ID: id})
	if err != nil {
		t.Fatalf("unable to GenerateHTTPResponse(%q, %d) test: %s", id, c, err)
	}
	if c == 0 {
		c = http.StatusCreated
	}
	return &http.Response{
		Body:       ioutil.NopCloser(bytes.NewBuffer(rb)),
		StatusCode: c,
	}
}
