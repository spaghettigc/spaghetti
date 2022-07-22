package main

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

func main() {
	http.HandleFunc("/api", func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello from a HandleFunc #1!\n")
	})

	http.HandleFunc("/file", func(w http.ResponseWriter, req *http.Request) {
		body, _ := ioutil.ReadAll(req.Body)

		// body := "TGVkZ2VyIElkLEFjY291bnQgTmFtZSxNYW5kYXRlIENvdW50cnksTWFuZGF0ZSBSZWZlcmVuY2UsTWFuZGF0ZSBTdGF0dXMsU2NoZW1lLE1hbmRhdGUgSWQsSWQKNzA1MTIsVUFUNDgsR0IsQklMTElOR1BMQVRGLVNVQVQ0OCxBQ1RJVkUsQmFjcyxNRDIwMjIwNjE1VUFUNDgsNzA1MTIK"

		dec, _ := base64.StdEncoding.DecodeString(string(body))
		reader := csv.NewReader(bytes.NewBuffer(dec))

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}

			fmt.Printf("first: %s second %s\n", record[0], record[1])
		}
		fmt.Println(string(dec))
	})

	http.HandleFunc("/list", func(w http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			w.Write([]byte(`[{
				"id": 1,
				"name": "Hanson Deck",
				"email": "hanson@deck.com",
				"sales": 37
			  }, {
				"id": 2,
				"name": "Sue Shei",
				"email": "sueshei@example.com",
				"sales": 550
			  }, {
				"id": 3,
				"name": "Jason Response",
				"email": "jason@response.com",
				"sales": 55
			  }, {
				"id": 4,
				"name": "Cher Actor",
				"email": "cher@example.com",
				"sales": 424
			  }, {
				"id": 5,
				"name": "Erica Widget",
				"email": "erica@widget.org",
				"sales": 243
			  }]`))

		case http.MethodPost:
			body, _ := ioutil.ReadAll(req.Body)
			bodyString := string(body)
			fmt.Println(bodyString)

			time.Sleep(time.Second * 3)
			io.WriteString(w, fmt.Sprintf("%s %s", bodyString, time.Now()))

		default:
			return
		}

	})

	http.ListenAndServe("localhost:3001", nil)
}
