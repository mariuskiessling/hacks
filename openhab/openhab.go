package openhab

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Client struct {
	HTTPClient http.Client
	APIBaseURL string
}

type Item struct {
	Link             string `json:"link"`
	State            string `json:"state"`
	StateDescription struct {
		Step     float64       `json:"step"`
		Pattern  string        `json:"pattern"`
		ReadOnly bool          `json:"readOnly"`
		Options  []interface{} `json:"options"`
	} `json:"stateDescription"`
	Editable   bool          `json:"editable"`
	Type       string        `json:"type"`
	Name       string        `json:"name"`
	Label      string        `json:"label"`
	Tags       []interface{} `json:"tags"`
	GroupNames []interface{} `json:"groupNames"`
}

func (c *Client) GetItem(itemName string) (item Item, err error) {
	resp, err := c.HTTPClient.Get(c.APIBaseURL + "/items/" + itemName)
	if err != nil {
		return Item{}, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Item{}, err
	}

	decoded := &Item{}
	err = json.Unmarshal(body, decoded)
	if err != nil {
		return Item{}, err
	}

	return *decoded, nil
}
