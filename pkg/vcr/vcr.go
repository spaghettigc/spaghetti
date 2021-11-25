package vcr

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"time"
)

type Request struct {
	// Body of request
	Body string `json:"body"`

	// Request headers
	Headers http.Header `json:"headers"`

	// Request URL
	URL string `json:"url"`

	// Request method
	Method string `json:"method"`

	Timestamp string `json:timestamp`

	Name string `json:name`
}

func now() string {
	return time.Now().Format(time.RFC3339)
}

func RequestHandler(r *http.Request, i interface{}, name string) (*Request, error) {
	// Copy the original request, so we can read the form values

	reqBytes, err := httputil.DumpRequest(r, true)
	if err != nil {
		panic(err)
		return nil, err
	}

	reqBuffer := bytes.NewBuffer(reqBytes)
	copiedReq, err := http.ReadRequest(bufio.NewReader(reqBuffer))
	if err != nil {
		panic(err)
		return nil, err
	}

	err = copiedReq.ParseForm()
	if err != nil {
		panic(err)
		return nil, err
	}

	err = json.NewDecoder(copiedReq.Body).Decode(&i)
	if err != nil {
		panic(err)
		return nil, err
	}

	bodyString, err := json.Marshal(i)
	if err != nil {
		panic(err)
		return nil, err
	}

	request := &Request{
		Body:      string(bodyString),
		Headers:   r.Header,
		URL:       r.URL.String(),
		Method:    r.Method,
		Timestamp: now(),
		Name:      name,
	}
	requestJSON, err := json.MarshalIndent(&request, "", "\t")

	if err != nil {
		panic(err)
		return nil, err
	}

	err = ioutil.WriteFile(fmt.Sprintf("recording/%s.json", name), requestJSON, 0644)

	if err != nil {
		panic(err)
		return nil, err
	}

	return request, nil
}

func ReadRequest(filename string) (Request, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return Request{}, err
	}

	r := Request{}

	err = json.Unmarshal(file, &r)
	if err != nil {
		panic(err)
		return r, err
	}

	return r, nil
}
