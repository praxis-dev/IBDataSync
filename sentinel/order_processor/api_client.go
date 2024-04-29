package orderprocessor

import (
	"io"
	"log"
	"net/http"
	"net/url"

	httpclient "sentinel/http_client"
)

type APIClient struct {
    apiToken   string
    apiURL     string
    httpClient httpclient.Client
}

func NewAPIClient(apiToken, apiURL string, httpClient httpclient.Client) *APIClient {
    return &APIClient{
        apiToken:   apiToken,
        apiURL:     apiURL,
        httpClient: httpClient,
    }
}

func (ac *APIClient) SendAPIRequest(data []byte) {
    jsonData := string(data)
    log.Printf("Sending data: %s\n", jsonData)

    form := url.Values{}
    form.Add("api_token", ac.apiToken)
    form.Add("data", jsonData)

    resp, err := ac.httpClient.PostForm(ac.apiURL, form)
    if err != nil {
        log.Fatalf("Error sending POST request: %v", err)
        return
    }
    defer resp.Body.Close()

    bodyBytes, err := io.ReadAll(resp.Body)
    if err != nil {
        log.Printf("Error reading response body: %v", err)
        return
    }
    body := string(bodyBytes)

    if resp.StatusCode != http.StatusOK {
        log.Printf("Received non-OK response: %d, response body: %s", resp.StatusCode, body)
    } else {
        log.Println("Data sent successfully and received OK response from the server.")
    }
}