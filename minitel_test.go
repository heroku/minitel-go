package minitel

import "testing"

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
