package tencent_vector_db

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type mockVectorClient struct {
	BaseURL    string
	HTTPClient *http.Client
	err        error
}

func (m *mockVectorClient) internalSendRequest(method, path string, data interface{}) error {
	return m.err
}

func Test_internalPrepareWriteRequest(t *testing.T) {
	tests := []struct {
		name     string
		vectors  []VectorData
		options  WriteOptions
		expected map[string]interface{}
	}{
		{
			name: "WithPartitionName",
			vectors: []VectorData{
				{
					ID:       "id1",
					Vector:   []float32{1.1, 2.2},
					Metadata: map[string]interface{}{"key": "value"},
				},
			},
			options: WriteOptions{
				CollectionName: "test_collection",
				PartitionName:  "test_partition",
			},
			expected: map[string]interface{}{
				"collection_name": "test_collection",
				"partition_name":  "test_partition",
				"vectors": []map[string]interface{}{
					{
						"id":       "id1",
						"vector":   []float32{1.1, 2.2},
						"metadata": map[string]interface{}{"key": "value"},
					},
				},
			},
		},
		{
			name: "WithoutPartitionName",
			vectors: []VectorData{
				{
					ID:     "id2",
					Vector: []float32{3.3},
				},
			},
			options: WriteOptions{
				CollectionName: "another_collection",
			},
			expected: map[string]interface{}{
				"collection_name": "another_collection",
				"vectors": []map[string]interface{}{
					{
						"id":     "id2",
						"vector": []float32{3.3},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := internalPrepareWriteRequest(tt.vectors, tt.options)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestWriteVectors(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := &VectorClient{
			BaseURL:    server.URL,
			HTTPClient: server.Client(),
		}

		vectors := []VectorData{{ID: "test_id", Vector: []float32{1.0}}}
		options := WriteOptions{CollectionName: "test_coll"}

		err := WriteVectors(client, vectors, options)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("ClientError", func(t *testing.T) {
		mockClient := &mockVectorClient{err: errors.New("connection error")}
		vectors := []VectorData{{ID: "test_id", Vector: []float32{1.0}}}
		options := WriteOptions{CollectionName: "test_coll"}

		err := WriteVectors(mockClient, vectors, options)
		if err == nil || err.Error() != "connection error" {
			t.Errorf("expected 'connection error', got %v", err)
		}
	})
}
