package main

import (
	"net/http"
	"testing"
)

func TestRequestStore(t *testing.T) {
	store := RequestStore{}

	buildAndStore := func(method, path string) *http.Request {
		req, err := http.NewRequest(method, path, nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := store.Save(req); err != nil {
			t.Fatal(err)
		}

		return req
	}

	get := buildAndStore("GET", "/path-a")
	post := buildAndStore("POST", "/path-b")

	found, err := store.Where(func(r *http.Request) (bool, error) {
		return r.Method == "GET", nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if len(found) != 1 {
		t.Errorf("expected 1 request to be found, got: %d", found)
	}

	if get.URL.Path != found[0].URL.Path {
		t.Errorf("unexpected request found: %#v", found)
	}

	found, err = store.Where(func(r *http.Request) (bool, error) {
		return r.Method == "POST", nil
	})

	if err != nil {
		t.Fatal(err)
	}

	if len(found) != 1 {
		t.Errorf("expected 1 request to be found, got: %d", found)
	}

	if post.URL.Path != found[0].URL.Path {
		t.Errorf("unexpected request found: %#v", found)
	}
}
