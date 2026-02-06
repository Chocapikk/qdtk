package lib

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	Transport: &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
	},
	Timeout: 30 * time.Second,
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

type QdrantClient struct {
	BaseUrl   string
	LogOutput io.Writer
	ApiKey    string
}

func NewQdrantClient(baseUrl string) *QdrantClient {
	return &QdrantClient{
		BaseUrl:   baseUrl,
		LogOutput: io.Discard,
	}
}

func (c *QdrantClient) Request(method, path string, body interface{}, target interface{}) error {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseUrl+path, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "qdtk/1.0.0 (+https://github.com/chocapikk/qdtk)")
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.ApiKey != "" {
		req.Header.Set("api-key", c.ApiKey)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return fmt.Errorf("authentication required (HTTP %d)", resp.StatusCode)
	}

	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if target != nil {
		return json.NewDecoder(resp.Body).Decode(target)
	}
	return nil
}

// Qdrant API response types

type QdrantResponse[T any] struct {
	Result T      `json:"result"`
	Status string `json:"status"`
	Time   float64 `json:"time"`
}

type CollectionsList struct {
	Collections []CollectionDescription `json:"collections"`
}

type CollectionDescription struct {
	Name string `json:"name"`
}

type CollectionInfo struct {
	Status          string `json:"status"`
	OptimizerStatus string `json:"optimizer_status"`
	VectorsCount    int64  `json:"vectors_count"`
	IndexedVectors  int64  `json:"indexed_vectors_count"`
	PointsCount     int64  `json:"points_count"`
	SegmentsCount   int    `json:"segments_count"`
	Config          struct {
		Params struct {
			Vectors interface{} `json:"vectors"`
		} `json:"params"`
	} `json:"config"`
	PayloadSchema map[string]interface{} `json:"payload_schema"`
}

type ScrollRequest struct {
	Limit       int     `json:"limit"`
	WithPayload bool    `json:"with_payload"`
	WithVector  bool    `json:"with_vector"`
	Offset      *string `json:"offset,omitempty"`
}

type ScrollResponse struct {
	Points         []Point     `json:"points"`
	NextPageOffset interface{} `json:"next_page_offset"`
}

type Point struct {
	Id      interface{}            `json:"id"`
	Payload map[string]interface{} `json:"payload"`
	Vector  interface{}            `json:"vector,omitempty"`
}

type SearchRequest struct {
	Vector      []float64 `json:"vector,omitempty"`
	Filter      *Filter   `json:"filter,omitempty"`
	Limit       int       `json:"limit"`
	WithPayload bool      `json:"with_payload"`
	WithVector  bool      `json:"with_vector"`
}

type Filter struct {
	Must    []Condition `json:"must,omitempty"`
	Should  []Condition `json:"should,omitempty"`
	MustNot []Condition `json:"must_not,omitempty"`
}

type Condition struct {
	Key   string      `json:"key,omitempty"`
	Match interface{} `json:"match,omitempty"`
}

type ClusterInfo struct {
	Status string `json:"status"`
	PeerId int64  `json:"peer_id"`
}

type TelemetryInfo struct {
	App struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"app"`
	Collections struct {
		NumberOfCollections int `json:"number_of_collections"`
	} `json:"collections"`
}
