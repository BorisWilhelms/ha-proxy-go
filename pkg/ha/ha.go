package ha

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

type HomeAssistant struct {
	BaseUrl     string
	AccessToken string

	client *http.Client
}

func NewHomeAssistant(baseUrl string, accessToken string) *HomeAssistant {
	return &HomeAssistant{
		BaseUrl:     baseUrl,
		AccessToken: accessToken,
		client:      &http.Client{},
	}
}

func (ha *HomeAssistant) CallService(service string, action string, entity string) bool {
	data, _ := json.Marshal(map[string]string{"entity_id": entity})
	reader := bytes.NewReader(data)
	req := ha.createRequest(http.MethodPost, fmt.Sprintf("services/%s/%s", service, action), reader)

	res, err := ha.doCall(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()
	return res.StatusCode == 200
}

func (ha *HomeAssistant) GetState(entityId string) Entity {
	req := ha.createRequest(http.MethodGet, fmt.Sprintf("states/%s", entityId), nil)
	res, err := ha.doCall(req)
	if err != nil {
		return Entity{}
	}
	defer res.Body.Close()

	var e Entity
	err = json.NewDecoder(res.Body).Decode(&e)
	if err != nil {
		log.Println("GetState: error parsing response", err)
		return Entity{}
	}

	return e
}

func (ha *HomeAssistant) createRequest(method string, requestUrl string, body io.Reader) *http.Request {
	requestUrl = fmt.Sprintf("%s/api/%s", ha.BaseUrl, requestUrl)
	req, _ := http.NewRequest(method, requestUrl, body)
	bearer := fmt.Sprintf("Bearer %s", ha.AccessToken)
	req.Header.Set("Authorization", bearer)

	return req
}

func (ha *HomeAssistant) doCall(r *http.Request) (*http.Response, error) {
	res, err := ha.client.Do(r)
	if err != nil {
		log.Println("doCall: error making http request:", r.Method, r.URL, err)
		return nil, err
	}
	if res.StatusCode != 200 {
		log.Println("doCall: request not succesfull:", r.Method, r.URL, res.Status)
		return nil, errors.New("doCall: request not succesfull")
	}

	return res, nil
}

type Entity struct {
	Entity_id  string `json:"entity_id"`
	Attributes map[string]interface{}
}

func (e *Entity) FriendlyName() string {
	return e.Attributes["friendly_name"].(string)
}
