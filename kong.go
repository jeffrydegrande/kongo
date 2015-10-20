package kongo

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

func NewKong(kongUrl string) *Kong {
	return &Kong{kongUrl}
}

func (kong *Kong) GetEndpoints() ([]Endpoint, error) {
	url := fmt.Sprintf("%s/apis", kong.Url)

	body, err := doGetRequest(url)
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

func (kong *Kong) SetEndpoint(endpoint *Endpoint) error {
	body, err := json.Marshal(endpoint)
	if err != nil {
		return err
	}

	endpointUrl := fmt.Sprintf("%s/apis/%s", kong.Url, endpoint.Name)
	status, err := doRequestWithBody("PATCH", endpointUrl, body)
	if err != nil {
		return err
	}

	if status >= 200 && status <= 300 {
		return nil
	}

	endpointUrl = fmt.Sprintf("%s/apis", kong.Url)
	status, err = doRequestWithBody("POST", endpointUrl, body)
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
	body, err := doGetRequest(url)
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

func (kong *Kong) SetPlugin(endpointNameOrId string, plugin string, config map[string]interface{}) error {
	pluginUrl := fmt.Sprintf("%s/apis/%s/plugins", kong.Url, endpointNameOrId)
	jsonPluginConfig, err := json.Marshal(config)

	status, err := doRequestWithBody("POST", pluginUrl, jsonPluginConfig)
	if err != nil {
		return err
	}

	if status < 400 {
		return nil
	}
	fmt.Printf("Status was %d, will update instead\n", status)

	status, err = doRequestWithBody("PATCH", pluginUrl, jsonPluginConfig)
	fmt.Printf("Status was %d\n", status)
	return err
}

func doGetRequest(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func doRequestWithBody(method string, url string, body []byte) (int, error) {
	client := &http.Client{}
	r, _ := http.NewRequest(method, url, bytes.NewBufferString(string(body)))
	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Content-Length", strconv.Itoa(len(string(body))))

	resp, _ := client.Do(r)

	defer resp.Body.Close()
	_, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return -1, err
	}
	return resp.StatusCode, nil
}
