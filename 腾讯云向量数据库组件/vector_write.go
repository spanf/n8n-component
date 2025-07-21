package tencent_vector_db

type VectorData struct {
	ID       string                 
	Vector   []float32              
	Metadata map[string]interface{} 
}

type WriteOptions struct {
	CollectionName string 
	PartitionName  string 
}

func WriteVectors(client *VectorClient, vectors []VectorData, options WriteOptions) error {
	requestData := internalPrepareWriteRequest(vectors, options)
	err := client.internalSendRequest("POST", "/vector/write", requestData)
	if err != nil {
		return err
	}
	return nil
}

func internalPrepareWriteRequest(vectors []VectorData, options WriteOptions) map[string]interface{} {
	request := make(map[string]interface{})
	request["collection_name"] = options.CollectionName
	if options.PartitionName != "" {
		request["partition_name"] = options.PartitionName
	}

	formattedVectors := make([]map[string]interface{}, len(vectors))
	for i, v := range vectors {
		formattedVectors[i] = map[string]interface{}{
			"id":       v.ID,
			"vector":   v.Vector,
			"metadata": v.Metadata,
		}
	}
	request["vectors"] = formattedVectors
	return request
}
