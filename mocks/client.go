package mocks

import (
	"errors"
	"sync"

	"github.com/heroku/minitel-go"
	"github.com/pborman/uuid"
)

// ErrorReporter is the interface necessary to report Errors.
// testing.T implements this interface.
type ErrorReporter interface {
	Errorf(string, ...interface{})
}

type response struct {
	result minitel.Result
	err    error
}

// MockClient implements minitel.Client
// Using the additional FollowupAndExpect..., and NotifyAndExpect..., one can
// mock successful and erroneous calls to the methods to ensure code handles
// these cases appropriately.
type MockClient struct {
	t                    ErrorReporter
	notifyExpectations   []response
	followupExpectations []response

	m sync.Mutex
}

var (
	errNoMoreExpectations = errors.New("no more expectations")
)

// NewMockClient returns a client that satisfies minitel.Client, but provides additional
// methods to be used for setting expectations around calls to Notify and Followup.
func NewMockClient(t ErrorReporter) *MockClient {
	return &MockClient{
		t: t,
	}
}

// Notify will succeed if there is a notify expectation waiting, otherwise it will fail
func (c *MockClient) Notify(p minitel.Payload) (result minitel.Result, err error) {
	c.m.Lock()
	defer c.m.Unlock()
	if len(c.notifyExpectations) > 0 {
		next := c.notifyExpectations[0]
		c.notifyExpectations = c.notifyExpectations[1:]
		return next.result, next.err
	}
	c.t.Errorf("no more notify expectations")
	return minitel.Result{}, errNoMoreExpectations
}

// Followup will succeed if there is a followup expectation waiting, otherwise it will fail
func (c *MockClient) Followup(id, body string) (result minitel.Result, err error) {
	c.m.Lock()
	defer c.m.Unlock()

	if len(c.followupExpectations) > 0 {
		next := c.followupExpectations[0]
		c.followupExpectations = c.followupExpectations[1:]
		return next.result, next.err
	}

	c.t.Errorf("no more followup expectations")
	return minitel.Result{}, errNoMoreExpectations
}

// NotifyAndExpectSuccess sets an expectation that Notify will be called, and will succeed
func (c *MockClient) NotifyAndExpectSuccess() string {
	c.m.Lock()
	defer c.m.Unlock()

	resultUUID := uuid.New()
	c.notifyExpectations = append(c.notifyExpectations, response{result: minitel.Result{Id: resultUUID}})
	return resultUUID
}

// NotifyAndExpectFailure sets an expectation that Notify will be called, and will not succeed
func (c *MockClient) NotifyAndExpectFailure(err error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.notifyExpectations = append(c.notifyExpectations, response{err: err})
}

// FollowupAndExpectSuccess sets an expectation that Followup will be called, and will succeed
func (c *MockClient) FollowupAndExpectSuccess() string {
	c.m.Lock()
	defer c.m.Unlock()

	resultUUID := uuid.New()
	c.followupExpectations = append(c.followupExpectations, response{result: minitel.Result{Id: resultUUID}})
	return resultUUID
}

// FollowupAndExpectFailure sets an expectation that Followup will be called, and will not succeed
func (c *MockClient) FollowupAndExpectFailure(err error) {
	c.m.Lock()
	defer c.m.Unlock()

	c.followupExpectations = append(c.followupExpectations, response{err: err})
}

// ExpectDone reports an error if there are pending expectations.
func (c *MockClient) ExpectDone() {
	c.m.Lock()
	defer c.m.Unlock()

	if len(c.notifyExpectations) > 0 {
		c.t.Errorf("%d Notify expectations left", len(c.notifyExpectations))
	}
	if len(c.followupExpectations) > 0 {
		c.t.Errorf("%d Followup expectations left", len(c.followupExpectations))
	}
}
