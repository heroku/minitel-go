package mocks

import (
	"errors"

	"github.com/heroku/minitel-go"
	"github.com/pborman/uuid"
)

type ErrorReporter interface {
	Errorf(string, ...interface{})
}

type MockClient struct {
	t                    ErrorReporter
	notifyExpectations   []error
	followupExpectations []error
}

var (
	errNoMoreExpectations = errors.New("no more expecations")
)

func NewMockClient(t ErrorReporter) (*MockClient, error) {
	return &MockClient{
		t:                    t,
		notifyExpectations:   make([]error, 0),
		followupExpectations: make([]error, 0),
	}, nil
}

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

func (c *MockClient) NotifyAndExpectSuccess() {
	c.notifyExpectations = append(c.notifyExpectations, nil)
}

func (c *MockClient) NotifyAndExpectFailure(err error) {
	c.notifyExpectations = append(c.notifyExpectations, err)
}

func (c *MockClient) FollowupAndExpectSuccess() {
	c.followupExpectations = append(c.followupExpectations, nil)
}

func (c *MockClient) FollowupAndExpectFailure(err error) {
	c.followupExpectations = append(c.followupExpectations, err)
}
