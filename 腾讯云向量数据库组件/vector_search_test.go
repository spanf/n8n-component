package tencent_vector_db_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"tencent_vector_db"
	
	"github.com/stretchr/testify/assert"
)

type mockVectorClient struct {
	baseURL string
}

func (c *mockVectorClient) internalSendRequest(method, path string, body map[string]interface{}) ([]byte, error) {
	return nil, nil
}

func TestSearchVectors_Success(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求路径
		assert.Equal(t, "/search", r.URL.Path)
		
		// 验证请求体
		var body map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&body)
		assert.NoError(t, err)
		assert.Equal(t, "test-collection", body["collection_name"])
		assert.Equal(t, []interface{}{1.0, 2.0}, body["vector"])
		assert.Equal(t, 5.0, body["top_k"])
		assert.Equal(t, "test-partition", body["partition_name"])
		assert.Equal(t, "color=red", body["filter"])

		// 返回成功响应
		response := tencent_vector_db.apiResponse{
			Code:    0,
			Message: "success",
			Data: []struct {
				ID       string                 `json:"id"`
				Score    float32                `json:"score"`
				Vector   []float32              `json:"vector"`
				Metadata map[string]interface{} `json:"metadata"`
			}{
				{
					ID:     "vec1",
					Score:  0.95,
					Vector: []float32{0.1, 0.2},
					Metadata: map[string]interface{}{
						"color": "red",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// 创建客户端
	client := &tencent_vector_db.VectorClient{}

	// 设置选项
	options := tencent_vector_db.SearchOptions{
		CollectionName: "test-collection",
		PartitionName:  "test-partition",
		TopK:           5,
		Filter:         "color=red",
	}

	// 执行搜索
	results, err := tencent_vector_db.SearchVectors(client, []float32{1, 2}, options)

	// 验证结果
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "vec1", results[0].ID)
	assert.Equal(t, float32(0.95), results[0].Score)
	assert.Equal(t, []float32{0.1, 0.2}, results[0].Vector)
	assert.Equal(t, "red", results[0].Metadata["color"])
}

func TestSearchVectors_APIFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := tencent_vector_db.apiResponse{
			Code:    1001,
			Message: "collection not found",
			Data:    nil,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &tencent_vector_db.VectorClient{}
	options := tencent_vector_db.SearchOptions{CollectionName: "invalid-collection"}

	results, err := tencent_vector_db.SearchVectors(client, []float32{1, 2}, options)

	assert.Nil(t, results)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "API error 1001")
}

func TestSearchVectors_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{invalid json}"))
	}))
	defer server.Close()

	client := &tencent_vector_db.VectorClient{}
	options := tencent_vector_db.SearchOptions{CollectionName: "test"}

	results, err := tencent_vector_db.SearchVectors(client, []float32{1, 2}, options)

	assert.Nil(t, results)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "JSON decode failed")
}

func TestSearchVectors_OptionalFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求体
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		
		// 确保可选字段不存在
		_, hasPartition := body["partition_name"]
		_, hasFilter := body["filter"]
		assert.False(t, hasPartition)
		assert.False(t, hasFilter)

		// 返回空成功响应
		response := tencent_vector_db.apiResponse{Code: 0}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &tencent_vector_db.VectorClient{}
	options := tencent_vector_db.SearchOptions{
		CollectionName: "minimal",
		TopK:           3,
	}

	_, err := tencent_vector_db.SearchVectors(client, []float32{3, 4}, options)
	assert.NoError(t, err)
}

func TestSearchVectors_ZeroResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := tencent_vector_db.apiResponse{
			Code:    0,
			Message: "success",
			Data:    []struct{}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &tencent_vector_db.VectorClient{}
	options := tencent_vector_db.SearchOptions{CollectionName: "empty"}

	results, err := tencent_vector_db.SearchVectors(client, []float32{1, 2}, options)

	assert.NoError(t, err)
	assert.Empty(t, results)
}
