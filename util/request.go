package util

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

func SendHttpRequest(method, url string, body []byte) ([]byte, error) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	url = strings.Trim(url, " ")
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	log.Println(fmt.Sprintf("Sending %v request to %v with body %s", method, url, string(body)))
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, errors.New(resp.Status)
	}
	response, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	log.Println(fmt.Sprintf("Getting %v response from %v", method, url))
	return response, err
}
