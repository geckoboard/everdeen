package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/satori/go.uuid"
)

type RequestStore struct {
	requestCount int
	mutex        sync.RWMutex
}

func (rs *RequestStore) Save(expUuid uuid.UUID, r *http.Request) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()

	rs.requestCount += 1

	//Ensure directories exist
	expPath := path.Join(*requestBaseStore, expUuid.String())
	newFileName := strconv.Itoa(rs.requestCount) + ".json"
	os.MkdirAll(expPath, 0744)

	var b []byte
	var err error

	if r.Body != nil {
		defer r.Body.Close()

		b, err = ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		r.Body = ioutil.NopCloser(bytes.NewReader(b))
	}

	request := Request{
		URL:        r.URL.String(),
		Method:     r.Method,
		Headers:    r.Header,
		BodyBase64: base64.StdEncoding.EncodeToString(b),
	}

	reqJson, err := json.Marshal(&request)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(expPath, newFileName), reqJson, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (rs *RequestStore) Where(expUuid uuid.UUID) ([]Request, error) {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()

	found := []Request{}

	basePath := path.Join(*requestBaseStore, expUuid.String())
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		// If the expectation directory doesn't exist should return empty array
		return found, nil
	}

	for _, file := range files {
		var data Request

		fBytes, err := ioutil.ReadFile(path.Join(basePath, file.Name()))
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(fBytes, &data)
		if err != nil {
			return nil, err
		}

		found = append(found, data)
	}

	return found, nil
}

func (s *Server) findExpectationByUuid(expUuid uuid.UUID) *Expectation {
	for _, exp := range s.expectations {
		if uuid.Equal(exp.Uuid, expUuid) {
			return exp
		}
	}

	return nil
}
