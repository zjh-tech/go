package enet

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

func PostJson(url string, js interface{}) (content []byte, err error) {
	datas, marshalErr := json.Marshal(js)
	if marshalErr != nil {
		ELog.ErrorAf("PostJson json.Marshal Error %v", marshalErr)
		return nil, marshalErr
	}

	return Post(url, datas)
}

func Post(url string, data []byte) (content []byte, err error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return content, err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 3 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return content, err
	}
	defer res.Body.Close()
	body, bodyError := ioutil.ReadAll(res.Body)
	return body, bodyError
}

var GSingleHttpConn IHttpConnection //Single Logic
var GMultiHttpConn IHttpConnection  //Multi  Logic
