package tencent_vector_db

import (
	"encoding/json"
	"fmt"
)

type SearchOptions struct {
	CollectionName string
	PartitionName  string
	TopK           int
	Filter         string
}

type SearchResult struct {
	ID       string
	Score    float32
	Vector   []float32
	Metadata map[string]interface{}
}

type apiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		ID       string                 `json:"id"`
		Score    float32                `json:"score"`
		Vector   []float32              `json:"vector"`
		Metadata map[string]interface{} `json:"metadata"`
	} `json:"data"`
}

func SearchVectors(client *VectorClient, queryVector []float32, options SearchOptions) ([]SearchResult, error) {
	requestBody := internalPrepareSearchRequest(queryVector, options)
	response, err := client.internalSendRequest("POST", "/search", requestBody)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	var apiResp apiResponse
	if err := json.Unmarshal(response, &apiResp); err != nil {
		return nil, fmt.Errorf("JSON decode failed: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, fmt.Errorf("API error %d: %s", apiResp.Code, apiResp.Message)
	}

	results := make([]SearchResult, len(apiResp.Data))
	for i, item := range apiResp.Data {
		results[i] = SearchResult{
			ID:       item.ID,
			Score:    item.Score,
			Vector:   item.Vector,
			Metadata: item.Metadata,
		}
	}
	return results, nil
}

func internalPrepareSearchRequest(queryVector []float32, options SearchOptions) map[string]interface{} {
	req := map[string]interface{}{
		"collection_name": options.CollectionName,
		"vector":          queryVector,
		"top_k":           options.TopK,
	}

	if options.PartitionName != "" {
		req["partition_name"] = options.PartitionName
	}
	if options.Filter != "" {
		req["filter"] = options.Filter
	}
	return req
}
