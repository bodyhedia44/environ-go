package main

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
)

// GenerateRecordKey generates a unique key for a record
func GenerateRecordKey(record map[string]interface{}) string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	recordJSON, _ := json.Marshal(record)
	hash := simpleHash(string(recordJSON))
	return fmt.Sprintf("PROOF_%d_%s", timestamp, hash)
}

// simpleHash generates a simple hash from a string
func simpleHash(str string) string {
	h := fnv.New32a()
	h.Write([]byte(str))
	hashValue := h.Sum32()
	return strconv.FormatUint(uint64(hashValue), 36)
}

// getAllResults collects all results from an iterator
func getAllResults(iterator shim.StateQueryIteratorInterface) ([]map[string]interface{}, error) {
	allResults := []map[string]interface{}{}

	for iterator.HasNext() {
		queryResponse, err := iterator.Next()
		if err != nil {
			return nil, err
		}

		var record interface{}
		err = json.Unmarshal(queryResponse.Value, &record)
		if err != nil {
			fmt.Printf("Error unmarshaling: %v\n", err)
			record = string(queryResponse.Value)
		}

		result := map[string]interface{}{
			"Key":    queryResponse.Key,
			"Record": record,
		}
		allResults = append(allResults, result)
	}

	return allResults, nil
}
