package api

import (
	"bytes"
	"encoding/json"
	"github.com/Labutin/MemoryKeyValueStorage/kvstorage"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	server *httptest.Server
)

type testResponse struct {
	responseCode int
	response     Resp
}

type testRequest struct {
	url      string
	method   string
	body     map[string]interface{}
	response testResponse
}

func TestPing(t *testing.T) {
	resp, err := http.Get(server.URL + "/v1/ping")
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, []byte("pong"), body)
}

func checkRequest(t *testing.T, testRequest testRequest, resp *http.Response, err error) {
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, testRequest.response.responseCode, resp.StatusCode)
	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	var response Resp
	err = json.Unmarshal(respBody, &response)
	require.NoError(t, err)
	require.Equal(t, testRequest.response.response, response)
}

func testRequests(t *testing.T, requests []testRequest) {
	for _, testRequest := range requests {
		switch testRequest.method {
		case "POST":
			body, err := json.Marshal(testRequest.body)
			require.NoError(t, err)
			resp, err := http.Post(testRequest.url, "application/json", bytes.NewBuffer(body))
			checkRequest(t, testRequest, resp, err)
		case "GET":
			resp, err := http.Get(testRequest.url)
			checkRequest(t, testRequest, resp, err)
		}
	}
}

func TestDict(t *testing.T) {
	nestedMap := map[string]interface{}{
		"t1": float64(1),
		"t2": float64(2),
	}
	postBody := map[string]interface{}{}
	postBody["key"] = "dict"
	postBody["value"] = kvstorage.Dict{
		"k1": float64(1),
		"k2": nestedMap,
	}
	requests := []testRequest{
		{
			url:    server.URL + urlPath + "/dict/",
			method: "POST",
			body:   postBody,
			response: testResponse{
				responseCode: 200,
				response: Resp{
					Response: "",
					Ok:       true,
				},
			},
		},
		{
			url:    server.URL + urlPath + "/getdict/dict/k2",
			method: "GET",
			response: testResponse{
				responseCode: 200,
				response: Resp{
					Response: nestedMap,
					Ok:       true,
				},
			},
		},
		{
			url:    server.URL + urlPath + "/getdict/dict/absent",
			method: "GET",
			response: testResponse{
				responseCode: 404,
				response: Resp{
					Error: "Key in dictionary not found",
					Ok:    false,
				},
			},
		},
	}
	testRequests(t, requests)
	value, ok := storage.Get("dict")
	require.True(t, ok)
	require.Equal(t, postBody["value"], value)
}

func TestAddRecord(t *testing.T) {
	postBodyt1 := map[string]interface{}{}
	postBodyt1["key"] = "t1"
	postBodyt1["value"] = "v1"
	postBodyt1["ttl"] = 0
	postBodyt2 := map[string]interface{}{}
	postBodyt2["key"] = "t2"
	postBodyt2["value"] = float64(2)
	postBodyt2["ttl"] = 0
	postBodyt3 := map[string]interface{}{}
	postBodyt3["key"] = "t3"
	postBodyt3["value"] = []interface{}{float64(0), float64(1), float64(2), float64(3)}
	postBodyt3["ttl"] = 0
	requests := []testRequest{
		{
			url:    server.URL + urlPath,
			method: "POST",
			body:   postBodyt1,
			response: testResponse{
				responseCode: 200,
				response: Resp{
					Response: "",
					Ok:       true,
				},
			},
		},
		{
			url:    server.URL + urlPath + "/get/t1",
			method: "GET",
			response: testResponse{
				responseCode: 200,
				response: Resp{
					Response: postBodyt1["value"],
					Ok:       true,
				},
			},
		},
		{
			url:    server.URL + urlPath,
			method: "POST",
			body:   postBodyt2,
			response: testResponse{
				responseCode: 200,
				response: Resp{
					Response: "",
					Ok:       true,
				},
			},
		},
		{
			url:    server.URL + urlPath + "/get/t2",
			method: "GET",
			response: testResponse{
				responseCode: 200,
				response: Resp{
					Response: postBodyt2["value"],
					Ok:       true,
				},
			},
		},
		{
			url:    server.URL + urlPath,
			method: "POST",
			body:   postBodyt3,
			response: testResponse{
				responseCode: 200,
				response: Resp{
					Response: "",
					Ok:       true,
				},
			},
		},
		{
			url:    server.URL + urlPath + "/get/t3",
			method: "GET",
			response: testResponse{
				responseCode: 200,
				response: Resp{
					Response: postBodyt3["value"],
					Ok:       true,
				},
			},
		},
		{
			url:    server.URL + urlPath + "/get/absent",
			method: "GET",
			response: testResponse{
				responseCode: 404,
				response: Resp{
					Response: nil,
					Error:    KeyNotFound.String(),
					Ok:       false,
				},
			},
		},
	}
	testRequests(t, requests)
	value, ok := storage.Get(postBodyt1["key"].(string))
	require.True(t, ok)
	require.Equal(t, postBodyt1["value"], value)
	value, ok = storage.Get(postBodyt2["key"].(string))
	require.True(t, ok)
	require.Equal(t, postBodyt2["value"], value)
	value, ok = storage.Get(postBodyt3["key"].(string))
	require.True(t, ok)
	require.Equal(t, postBodyt3["value"], value)
}

func init() {
	InitStorage(10)
	server = httptest.NewServer(InitRouter())
}
