package kongo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

type Kong struct {
	Url string
}

type EndpointResult struct {
	Data []Endpoint `json:"data"`
}

type Endpoint struct {
	ID           string `json:"id"`
	Path         string `json:"path"`
	TargetUrl    string `json:"target_url"`
	Name         string `json:"name"`
	CreatedAt    int64  `json:"created_at"`
	PreserveHost bool   `json:"preserve_host"`
	StripPath    bool   `json:"strip_path"`
}

type PluginResult struct {
	Data []Plugin `json:"data"`
}

type Plugin struct {
	ApiID     string                 `json:"api_id"`
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Value     map[string]interface{} `json:"value"`
	Enabled   bool                   `json:"enabled"`
	CreatedAt int64                  `json:"created_at"`
}

func _doGetRequest(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func NewKong(kongUrl string) *Kong {
	return &Kong{kongUrl}
}

func (kong *Kong) GetEndpoints() ([]Endpoint, error) {
	url := fmt.Sprintf("%s/apis", kong.Url)

	body, err := _doGetRequest(url)
	if err != nil {
		return nil, err
	}

	var result EndpointResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (kong *Kong) postEndpoint(endpoint *Endpoint) (int, error) {
	endpointUrl := fmt.Sprintf("%s/apis", kong.Url)
	data := url.Values{}
	data.Set("strip_path", strconv.FormatBool(endpoint.StripPath))
	data.Set("preserve_host", strconv.FormatBool(endpoint.PreserveHost))
	data.Set("name", endpoint.Name)
	data.Set("path", endpoint.Path)
	data.Set("target_url", endpoint.TargetUrl)

	client := &http.Client{}
	r, _ := http.NewRequest("POST", endpointUrl, bytes.NewBufferString(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, _ := client.Do(r)

	defer resp.Body.Close()
	_, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}
	return resp.StatusCode, nil
}

func (kong *Kong) patchEndpoint(endpoint *Endpoint) (int, error) {
	endpointUrl := fmt.Sprintf("%s/apis/%s", kong.Url, endpoint.Name)
	data := url.Values{}
	data.Set("strip_path", strconv.FormatBool(endpoint.StripPath))
	data.Set("preserve_host", strconv.FormatBool(endpoint.PreserveHost))
	data.Set("name", endpoint.Name)
	data.Set("path", endpoint.Path)
	data.Set("target_url", endpoint.TargetUrl)

	client := &http.Client{}
	r, _ := http.NewRequest("PATCH", endpointUrl, bytes.NewBufferString(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, _ := client.Do(r)
	defer resp.Body.Close()
	_, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}
	return resp.StatusCode, nil
}

func (kong *Kong) SetEndpoint(endpoint *Endpoint) error {
	status, err := kong.patchEndpoint(endpoint)
	if err != nil {
		return err
	}

	if status >= 200 && status <= 300 {
		return nil
	}

	status, err = kong.postEndpoint(endpoint)
	if err != nil {
		return err
	}

	return nil
}

func NewEndpoint(name string) *Endpoint {
	return &Endpoint{Name: name, Path: fmt.Sprintf("/%s", name)}
}

func (kong *Kong) GetPlugins(endpointNameOrId string) ([]Plugin, error) {
	url := fmt.Sprintf("%s/apis/%s/plugins", kong.Url, endpointNameOrId)
	body, err := _doGetRequest(url)
	if err != nil {
		return nil, err
	}

	var result PluginResult
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	return result.Data, nil
}
