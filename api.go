package trestCommon

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func GetApi(token, url string) ([]byte, error) {
	method := "GET"
	var bearer = "Bearer " + token
	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Add("Authorization", bearer)
	res, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}
func PostApi(token, url string, doc interface{}) ([]byte, error) {
	method := "POST"
	var bearer = "Bearer " + token
	requestByte, err := json.Marshal(doc)
	if err != nil {
		return []byte{}, err
	}
	requestReader := bytes.NewReader(requestByte)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, requestReader)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Add("Authorization", bearer)
	res, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}


func PostApiwithBasicAuth(auth string,url string,body interface{})([]byte, error){
	method := "POST"
	basicAuth := "Basic " + auth
	client := &http.Client{}
	requestByte, err := json.Marshal(body)
	if err != nil {
		return []byte{}, err
	}
	requestReader := bytes.NewReader(requestByte)
	req, err := http.NewRequest(method, url, requestReader)
	if err != nil {
		return []byte{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", basicAuth)
	res, err := client.Do(req)
	if err != nil {
		return []byte{}, err
	}
	defer res.Body.Close()
	return ioutil.ReadAll(res.Body)
}
