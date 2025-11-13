package main

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// WeightComparison handles weight comparison operations
type WeightComparison struct{}

// NewWeightComparison creates a new WeightComparison instance
func NewWeightComparison() *WeightComparison {
	return &WeightComparison{}
}

// ComparisonResult represents a single comparison result
type ComparisonResult struct {
	IncrementID    int     `json:"incrementId"`
	ChainedWeight  float64 `json:"chainedWeight"`
	ReceivedWeight float64 `json:"receivedWeight"`
}

// ComparisonResponse represents the response from weight comparison
type ComparisonResponse struct {
	Success        bool               `json:"success"`
	Results        []ComparisonResult `json:"results"`
	DeletedRecords []string           `json:"deletedRecords"`
	Message        string             `json:"message,omitempty"`
}

// GroupData represents grouped data for comparison
type GroupData struct {
	ChainedWeightSum  float64
	ReceivedWeightMax *float64
	TicketCount       int
	RecordIDs         []string
}

// CompareWeightsByPressIncrement compares weights by press increment
func (wc *WeightComparison) CompareWeightsByPressIncrement(ctx contractapi.TransactionContextInterface, deleteViolations string) (string, error) {
	fmt.Println("============= START : Compare Weights By Press Increment ===========")

	shouldDelete := deleteViolations == "true"

	proofQuery := map[string]interface{}{
		"selector": map[string]interface{}{
			"docType": "proofRecord",
		},
	}
	proofQueryString, _ := json.Marshal(proofQuery)
	proofIterator, err := ctx.GetStub().GetQueryResult(string(proofQueryString))
	if err != nil {
		response := ComparisonResponse{
			Success:        false,
			Message:        fmt.Sprintf("Error comparing weights: %v", err),
			Results:        []ComparisonResult{},
			DeletedRecords: []string{},
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}
	defer proofIterator.Close()

	proofRecords, err := getAllResults(proofIterator)
	if err != nil {
		response := ComparisonResponse{
			Success:        false,
			Message:        fmt.Sprintf("Error comparing weights: %v", err),
			Results:        []ComparisonResult{},
			DeletedRecords: []string{},
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	ticketQuery := map[string]interface{}{
		"selector": map[string]interface{}{
			"docType": "ticket",
		},
	}
	ticketQueryString, _ := json.Marshal(ticketQuery)
	ticketIterator, err := ctx.GetStub().GetQueryResult(string(ticketQueryString))
	if err != nil {
		response := ComparisonResponse{
			Success:        false,
			Message:        fmt.Sprintf("Error comparing weights: %v", err),
			Results:        []ComparisonResult{},
			DeletedRecords: []string{},
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}
	defer ticketIterator.Close()

	tickets, err := getAllResults(ticketIterator)
	if err != nil {
		response := ComparisonResponse{
			Success:        false,
			Message:        fmt.Sprintf("Error comparing weights: %v", err),
			Results:        []ComparisonResult{},
			DeletedRecords: []string{},
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	groupedResults := make(map[float64]*GroupData)

	for _, item := range proofRecords {
		proof := item["Record"].(map[string]interface{})
		recordID := item["Key"].(string)

		pressIncrement, hasPressIncrement := proof["press_increment"]
		chainedWeight, hasChainedWeight := proof["chained_weight"]

		if hasPressIncrement && pressIncrement != nil && hasChainedWeight {
			var incrementID float64
			switch v := pressIncrement.(type) {
			case float64:
				incrementID = v
			case int:
				incrementID = float64(v)
			default:
				continue
			}

			var weight float64
			switch w := chainedWeight.(type) {
			case float64:
				weight = w
			case int:
				weight = float64(w)
			default:
				continue
			}

			if _, exists := groupedResults[incrementID]; !exists {
				groupedResults[incrementID] = &GroupData{
					ChainedWeightSum:  0,
					ReceivedWeightMax: nil,
					TicketCount:       0,
					RecordIDs:         []string{},
				}
			}

			groupedResults[incrementID].ChainedWeightSum += weight
			groupedResults[incrementID].RecordIDs = append(groupedResults[incrementID].RecordIDs, recordID)
		}
	}

	for _, item := range tickets {
		ticket := item["Record"].(map[string]interface{})

		incrementIDVal, hasIncrementID := ticket["incrementId"]
		receivedWeightVal, hasReceivedWeight := ticket["receivedWeight"]

		if !hasIncrementID || !hasReceivedWeight {
			continue
		}

		var incrementID float64
		switch v := incrementIDVal.(type) {
		case float64:
			incrementID = v
		case int:
			incrementID = float64(v)
		default:
			continue
		}

		var receivedWeight float64
		switch w := receivedWeightVal.(type) {
		case float64:
			receivedWeight = w
		case int:
			receivedWeight = float64(w)
		default:
			continue
		}

		if _, exists := groupedResults[incrementID]; !exists {
			groupedResults[incrementID] = &GroupData{
				ChainedWeightSum:  0,
				ReceivedWeightMax: nil,
				TicketCount:       0,
				RecordIDs:         []string{},
			}
		}

		if groupedResults[incrementID].ReceivedWeightMax == nil {
			groupedResults[incrementID].ReceivedWeightMax = &receivedWeight
		} else {
			maxVal := math.Max(*groupedResults[incrementID].ReceivedWeightMax, receivedWeight)
			groupedResults[incrementID].ReceivedWeightMax = &maxVal
		}
		groupedResults[incrementID].TicketCount++
	}

	results := []ComparisonResult{}
	deletedRecords := []string{}

	for incrementID, group := range groupedResults {
		if group.ChainedWeightSum > 0 &&
			group.TicketCount > 0 &&
			group.ReceivedWeightMax != nil &&
			group.ChainedWeightSum > *group.ReceivedWeightMax {

			results = append(results, ComparisonResult{
				IncrementID:    int(incrementID),
				ChainedWeight:  math.Round(group.ChainedWeightSum*100) / 100,
				ReceivedWeight: math.Round(*group.ReceivedWeightMax*100) / 100,
			})

			if shouldDelete {
				for _, recordID := range group.RecordIDs {
					err := ctx.GetStub().DelState(recordID)
					if err == nil {
						deletedRecords = append(deletedRecords, recordID)
					}
				}
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].IncrementID < results[j].IncrementID
	})

	fmt.Println("============= END : Compare Weights By Press Increment ===========")

	response := ComparisonResponse{
		Success:        true,
		Results:        results,
		DeletedRecords: deletedRecords,
	}

	if !shouldDelete {
		response.DeletedRecords = []string{}
	}

	responseJSON, _ := json.Marshal(response)
	return string(responseJSON), nil
}

// CompareWeightsByStoreIncrement compares weights by store increment
func (wc *WeightComparison) CompareWeightsByStoreIncrement(ctx contractapi.TransactionContextInterface, deleteViolations string) (string, error) {
	fmt.Println("============= START : Compare Weights By Store Increment ===========")

	shouldDelete := deleteViolations == "true"

	proofQuery := map[string]interface{}{
		"selector": map[string]interface{}{
			"docType": "proofRecord",
		},
	}
	proofQueryString, _ := json.Marshal(proofQuery)
	proofIterator, err := ctx.GetStub().GetQueryResult(string(proofQueryString))
	if err != nil {
		response := ComparisonResponse{
			Success:        false,
			Message:        fmt.Sprintf("Error comparing weights: %v", err),
			Results:        []ComparisonResult{},
			DeletedRecords: []string{},
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}
	defer proofIterator.Close()

	proofRecords, err := getAllResults(proofIterator)
	if err != nil {
		response := ComparisonResponse{
			Success:        false,
			Message:        fmt.Sprintf("Error comparing weights: %v", err),
			Results:        []ComparisonResult{},
			DeletedRecords: []string{},
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	ticketQuery := map[string]interface{}{
		"selector": map[string]interface{}{
			"docType": "ticket",
		},
	}
	ticketQueryString, _ := json.Marshal(ticketQuery)
	ticketIterator, err := ctx.GetStub().GetQueryResult(string(ticketQueryString))
	if err != nil {
		response := ComparisonResponse{
			Success:        false,
			Message:        fmt.Sprintf("Error comparing weights: %v", err),
			Results:        []ComparisonResult{},
			DeletedRecords: []string{},
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}
	defer ticketIterator.Close()

	tickets, err := getAllResults(ticketIterator)
	if err != nil {
		response := ComparisonResponse{
			Success:        false,
			Message:        fmt.Sprintf("Error comparing weights: %v", err),
			Results:        []ComparisonResult{},
			DeletedRecords: []string{},
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	groupedResults := make(map[float64]*GroupData)

	for _, item := range proofRecords {
		proof := item["Record"].(map[string]interface{})
		recordID := item["Key"].(string)

		storeIncrement, hasStoreIncrement := proof["store_increment"]
		chainedWeight, hasChainedWeight := proof["chained_weight"]

		if hasStoreIncrement && storeIncrement != nil && hasChainedWeight {
			var incrementID float64
			switch v := storeIncrement.(type) {
			case float64:
				incrementID = v
			case int:
				incrementID = float64(v)
			default:
				continue
			}

			var weight float64
			switch w := chainedWeight.(type) {
			case float64:
				weight = w
			case int:
				weight = float64(w)
			default:
				continue
			}

			if _, exists := groupedResults[incrementID]; !exists {
				groupedResults[incrementID] = &GroupData{
					ChainedWeightSum:  0,
					ReceivedWeightMax: nil,
					TicketCount:       0,
					RecordIDs:         []string{},
				}
			}

			groupedResults[incrementID].ChainedWeightSum += weight
			groupedResults[incrementID].RecordIDs = append(groupedResults[incrementID].RecordIDs, recordID)
		}
	}

	for _, item := range tickets {
		ticket := item["Record"].(map[string]interface{})

		incrementIDVal, hasIncrementID := ticket["incrementId"]
		receivedWeightVal, hasReceivedWeight := ticket["receivedWeight"]

		if !hasIncrementID || !hasReceivedWeight {
			continue
		}

		var incrementID float64
		switch v := incrementIDVal.(type) {
		case float64:
			incrementID = v
		case int:
			incrementID = float64(v)
		default:
			continue
		}

		var receivedWeight float64
		switch w := receivedWeightVal.(type) {
		case float64:
			receivedWeight = w
		case int:
			receivedWeight = float64(w)
		default:
			continue
		}

		if _, exists := groupedResults[incrementID]; !exists {
			groupedResults[incrementID] = &GroupData{
				ChainedWeightSum:  0,
				ReceivedWeightMax: nil,
				TicketCount:       0,
				RecordIDs:         []string{},
			}
		}

		if groupedResults[incrementID].ReceivedWeightMax == nil {
			groupedResults[incrementID].ReceivedWeightMax = &receivedWeight
		} else {
			maxVal := math.Max(*groupedResults[incrementID].ReceivedWeightMax, receivedWeight)
			groupedResults[incrementID].ReceivedWeightMax = &maxVal
		}
		groupedResults[incrementID].TicketCount++
	}

	results := []ComparisonResult{}
	deletedRecords := []string{}

	for incrementID, group := range groupedResults {
		if group.ChainedWeightSum > 0 &&
			group.TicketCount > 0 &&
			group.ReceivedWeightMax != nil &&
			group.ChainedWeightSum > *group.ReceivedWeightMax {

			results = append(results, ComparisonResult{
				IncrementID:    int(incrementID),
				ChainedWeight:  math.Round(group.ChainedWeightSum*100) / 100,
				ReceivedWeight: math.Round(*group.ReceivedWeightMax*100) / 100,
			})

			if shouldDelete {
				for _, recordID := range group.RecordIDs {
					err := ctx.GetStub().DelState(recordID)
					if err == nil {
						deletedRecords = append(deletedRecords, recordID)
					}
				}
			}
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].IncrementID < results[j].IncrementID
	})

	fmt.Println("============= END : Compare Weights By Store Increment ===========")

	response := ComparisonResponse{
		Success:        true,
		Results:        results,
		DeletedRecords: deletedRecords,
	}

	if !shouldDelete {
		response.DeletedRecords = []string{}
	}

	responseJSON, _ := json.Marshal(response)
	return string(responseJSON), nil
}
