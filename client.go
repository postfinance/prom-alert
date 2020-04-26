package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type client struct {
	url string
	*http.Client
}

func (c client) post(a ...alert) error {
	d, err := json.Marshal(a)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewBuffer(d))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode > 399 {
		body, _ := ioutil.ReadAll(resp.Body)

		return fmt.Errorf("request failed with statuscode %s: %s", resp.Status, string(body))
	}

	return nil
}
