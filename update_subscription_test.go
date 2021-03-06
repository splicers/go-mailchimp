package mailchimp_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	mailchimp "github.com/splicers/go-mailchimp"
	"github.com/splicers/go-mailchimp/status"
	"github.com/stretchr/testify/assert"
)

func TestUpdateSubscriptionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(400)
		rw.Header().Set("Content-Type", "application/json")
		fmt.Fprint(rw, invalidMergeFieldsErrorResponse)
	}))
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	client, err := mailchimp.NewClient("the_api_key-us13", &http.Client{Transport: transport})
	assert.NoError(t, err)

	baseURL, _ := url.Parse("http://localhost/")
	client.SetBaseURL(baseURL)

	memberResponse, err := client.Subscribe("list_id", "john@reese.com", map[string]interface{}{})
	assert.Nil(t, memberResponse)
	assert.Equal(t, "Error 400 Invalid Resource (Your merge fields were invalid.)", err.Error())

	errResponse, ok := err.(*mailchimp.ErrorResponse)
	assert.True(t, ok)
	assert.Equal(t, "Invalid Resource", errResponse.Title)
	assert.Equal(t, 400, errResponse.Status)
	assert.Equal(t, "Your merge fields were invalid.", errResponse.Detail)
}

func TestUpdateSubscriptionMalformedError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.WriteHeader(500)
	}))
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	client, err := mailchimp.NewClient("the_api_key-us13", &http.Client{Transport: transport})
	assert.NoError(t, err)

	baseURL, _ := url.Parse("http://localhost/")
	client.SetBaseURL(baseURL)

	memberResponse, err := client.UpdateSubscription("list_id", "john@reese.com", map[string]interface{}{})
	assert.Nil(t, memberResponse)
	assert.Equal(t, "unexpected end of JSON input", err.Error())
}

func TestUpdateSubscription(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var pld struct {
			Email  string `json:"email_address"`
			Status string `json:"status"`
		}
		decoder := json.NewDecoder(req.Body)
		defer req.Body.Close()
		assert.NoError(t, decoder.Decode(&pld))

		// Test that we can override the email but there are some default attributes.
		assert.Equal(t, "another@email.com", pld.Email)
		assert.Equal(t, status.Subscribed, pld.Status)

		rw.WriteHeader(200)
		rw.Header().Set("Content-Type", "application/json")
		fmt.Fprint(rw, successResponse)
	}))
	defer server.Close()

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(server.URL)
		},
	}

	client, err := mailchimp.NewClient("the_api_key-us13", &http.Client{Transport: transport})
	assert.NoError(t, err)

	baseURL, _ := url.Parse("http://localhost/")
	client.SetBaseURL(baseURL)

	memberResponse, err := client.UpdateSubscription("list_id", "john@reese.com", map[string]interface{}{
		"email_address": "another@email.com",
	})
	assert.NoError(t, err)
	assert.Equal(t, "11bf13d1eb58116eba1de370b2bd796b", memberResponse.ID)
}
