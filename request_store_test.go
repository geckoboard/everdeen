package main

import (
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/satori/go.uuid"
)

func TestRequestStore(t *testing.T) {
	store := RequestStore{}

	expUuid := uuid.NewV4()

	buildAndStore := func(method, path string, body string) *http.Request {
		req, err := http.NewRequest(method, path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}

		if err := store.Save(expUuid, req); err != nil {
			t.Fatal(err)
		}

		return req
	}

	get := buildAndStore("GET", "/path-a", "")
	post := buildAndStore("POST", "/path-b", "Hello World")

	found, err := store.Where(uuid.NewV4())
	if err != nil {
		t.Fatal(err)
	}

	if len(found) != 0 {
		t.Errorf("expected 0 requests to be found, but got: %d", found)
	}

	found, err = store.Where(expUuid)
	if err != nil {
		t.Fatal(err)
	}

	if len(found) != 2 {
		t.Errorf("expected %d requests to be found, got: %d", 2, found)
	}

	if get.URL.String() != found[0].URL {
		t.Errorf("unexpected request returned: %#v", found)
	}

	if post.URL.String() != found[1].URL {
		t.Errorf("unexpected request returned: %#v", found)
	}

	expectedReq := Request{
		URL:        "/path/b",
		Method:     "POST",
		BodyBase64: "SGVsbG8gV29ybGQ=",
	}

	if reflect.DeepEqual(expectedReq, found[1]) {
		t.Errorf("Expected request %+v to match %+v but it didn't", expectedReq, found[1])
	}
}
