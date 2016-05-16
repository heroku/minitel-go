package mocks

import (
	"errors"

	"github.com/heroku/minitel-go"
	"github.com/pborman/uuid"
)

// ErrorReporter is the interface necessary to report Errors.
// testing.T implements this interface.
type ErrorReporter interface {
	Errorf(string, ...interface{})
}

// MockClient implements minitel.Client
// Using the additional FollowupAndExpect..., and NotifyAndExpect..., one can
// mock successful and erroneous calls to the methods to ensure code handles
// these cases appropriately.
type MockClient struct {
	t                    ErrorReporter
	notifyExpectations   []error
	followupExpectations []error
}

var (
	errNoMoreExpectations = errors.New("no more expecations")
)

// NewMockClient returns a client that satisfies minitel.Client, but provides additional
// methods to be used for setting expectations around calls to Notify and Followup.
func NewMockClient(t ErrorReporter) (*MockClient, error) {
	return &MockClient{
		t:                    t,
		notifyExpectations:   make([]error, 0),
		followupExpectations: make([]error, 0),
	}, nil
}

// Notify will succeed if there is a notify expectation waiting, otherwise it will fail
func (c *MockClient) Notify(p minitel.Payload) (result minitel.Result, err error) {
	if len(c.notifyExpectations) > 0 {
		next := c.notifyExpectations[0]
		c.notifyExpectations = c.notifyExpectations[1:]

		if next == nil {
			return minitel.Result{Id: uuid.New()}, nil
		}
		return minitel.Result{}, next
	}
	c.t.Errorf("no more notify expecations")
	return minitel.Result{}, errNoMoreExpectations
}

// Followup will succeed if there is a followup expectation waiting, otherwise it will fail
func (c *MockClient) Followup(id, body string) (result minitel.Result, err error) {
	if len(c.followupExpectations) > 0 {
		next := c.followupExpectations[0]
		c.followupExpectations = c.followupExpectations[1:]
		if next == nil {
			return minitel.Result{Id: uuid.New()}, nil
		}
		return minitel.Result{}, next
	}

	c.t.Errorf("no more followup expecations")
	return minitel.Result{}, errNoMoreExpectations
}

// NotifyAndExpectSuccess sets an expectation that Notify will be called, and will succeed
func (c *MockClient) NotifyAndExpectSuccess() {
	c.notifyExpectations = append(c.notifyExpectations, nil)
}

// NotifyAndExpectFailure sets an expectation that Notify will be called, and will not succeed
func (c *MockClient) NotifyAndExpectFailure(err error) {
	c.notifyExpectations = append(c.notifyExpectations, err)
}

// FollowupAndExpectSuccess sets an expectation that Followup will be called, and will succeed
func (c *MockClient) FollowupAndExpectSuccess() {
	c.followupExpectations = append(c.followupExpectations, nil)
}

// FollowupAndExpectFailure sets an expectation that Followup will be called, and will not succeed
func (c *MockClient) FollowupAndExpectFailure(err error) {
	c.followupExpectations = append(c.followupExpectations, err)
}

// ExpectDone reports an error if there are pending expectations.
func (c *MockClient) ExpectDone() {
	if len(c.notifyExpectations) > 0 {
		c.t.Errorf("%d Notify expectations left", len(c.notifyExpectations))
	}
	if len(c.followupExpectations) > 0 {
		c.t.Errorf("%d Followup expectations left", len(c.followupExpectations))
	}
}
