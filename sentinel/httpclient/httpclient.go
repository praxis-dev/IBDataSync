package httpclient

import (
	"net/http"
	"net/url"
	"strings"
)

type Client interface {
    PostForm(url string, data url.Values) (resp *http.Response, err error)
}

type httpClient struct{}

func NewClient() Client {
    return &httpClient{}
}

func (c *httpClient) PostForm(url string, data url.Values) (resp *http.Response, err error) {
    formData := strings.NewReader(data.Encode())
    req, err := http.NewRequest("POST", url, formData)
    if err != nil {
        return nil, err
    }
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

    client := &http.Client{}
    resp, err = client.Do(req)
    if err != nil {
        return nil, err
    }

    return resp, nil
}