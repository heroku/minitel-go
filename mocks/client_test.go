package mocks

import (
	"errors"
	"fmt"
	"testing"

	"github.com/heroku/minitel-go"
)

type testReporterMock struct {
	errors []string
}

func newTestReporterMock() *testReporterMock {
	return &testReporterMock{errors: make([]string, 0)}
}

func (trm *testReporterMock) Errorf(format string, args ...interface{}) {
	trm.errors = append(trm.errors, fmt.Sprintf(format, args...))
}

func TestSatisfiesInterface(t *testing.T) {
	var client interface{} = &MockClient{}
	if _, ok := client.(minitel.Client); !ok {
		t.Fatalf("MockClient cannot be used as minitel.Client")
	}
}

func TestNotifySuccess(t *testing.T) {
	r := newTestReporterMock()

	errFoo := errors.New("error foo")
	client := NewMockClient(r)
	client.NotifyAndExpectSuccess()
	client.NotifyAndExpectFailure(errFoo)

	res, err := client.Notify(minitel.Payload{})
	if err != nil {
		t.Errorf("Expected success, got failure = %q", err)
	} else if res.Id == "" {
		t.Errorf("Expected result with non empty Id")
	}

	res, err = client.Notify(minitel.Payload{})
	if err != errFoo {
		t.Errorf("Expected %q, got failure = %q", errFoo)
	} else if res.Id != "" {
		t.Errorf("Expected result with empty Id")
	}

	res, err = client.Notify(minitel.Payload{})
	if err != errNoMoreExpectations {
		t.Errorf("Expected no more expectations, got %q", err)
	}
}

func TestFollowupSuccess(t *testing.T) {
	r := newTestReporterMock()

	errFoo := errors.New("error foo")
	client := NewMockClient(r)
	client.FollowupAndExpectSuccess()
	client.FollowupAndExpectFailure(errFoo)

	res, err := client.Followup("id", "body")
	if err != nil {
		t.Errorf("Expected success, got failure = %q", err)
	} else if res.Id == "" {
		t.Errorf("Expected result with non empty Id")
	}

	res, err = client.Followup("id", "body")
	if err != errFoo {
		t.Errorf("Expected %q, got failure = %q", errFoo)
	} else if res.Id != "" {
		t.Errorf("Expected result with empty Id")
	}

	res, err = client.Followup("id", "body")
	if err != errNoMoreExpectations {
		t.Errorf("Expected no more expectations, got %q", err)
	}
}

func TestExpectDone(t *testing.T) {
	r := newTestReporterMock()
	client := NewMockClient(r)
	client.NotifyAndExpectSuccess()
	client.FollowupAndExpectSuccess()
	client.ExpectDone()

	if len(r.errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(r.errors))
	}
}
