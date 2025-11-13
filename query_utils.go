package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// QueryUtils handles query operations
type QueryUtils struct{}

// NewQueryUtils creates a new QueryUtils instance
func NewQueryUtils() *QueryUtils {
	return &QueryUtils{}
}

// QueryProofRecord queries a proof record by ID
func (qu *QueryUtils) QueryProofRecord(ctx contractapi.TransactionContextInterface, recordId string) (string, error) {
	recordAsBytes, err := ctx.GetStub().GetState(recordId)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}
	if recordAsBytes == nil || len(recordAsBytes) == 0 {
		return "", fmt.Errorf("Proof record %s does not exist", recordId)
	}
	return string(recordAsBytes), nil
}

// QueryAllProofRecords queries all proof records
func (qu *QueryUtils) QueryAllProofRecords(ctx contractapi.TransactionContextInterface) (string, error) {
	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"docType": "proofRecord",
		},
	}

	queryString, err := json.Marshal(query)
	if err != nil {
		return "", err
	}

	resultsIterator, err := ctx.GetStub().GetQueryResult(string(queryString))
	if err != nil {
		return "", err
	}
	defer resultsIterator.Close()

	allResults, err := getAllResults(resultsIterator)
	if err != nil {
		return "", err
	}

	resultsJSON, err := json.Marshal(allResults)
	if err != nil {
		return "", err
	}

	return string(resultsJSON), nil
}

// QueryRecordsByField queries records by a specific field
func (qu *QueryUtils) QueryRecordsByField(ctx contractapi.TransactionContextInterface, fieldName string, fieldValue string) (string, error) {
	fmt.Printf("============= START : Query Records By Field %s ===========\n", fieldName)

	var parsedValue interface{} = fieldValue

	numericFields := []string{"parent_increment", "store_increment", "press_increment", "chained_weight"}
	isNumeric := false
	for _, nf := range numericFields {
		if nf == fieldName {
			isNumeric = true
			break
		}
	}

	if isNumeric {
		floatVal, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			fmt.Printf("Error querying records by %s: Invalid %s: %s is not a number\n", fieldName, fieldName, fieldValue)
			return "[]", nil
		}
		parsedValue = floatVal
	}

	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"docType": "proofRecord",
			fieldName: parsedValue,
		},
	}

	queryString, err := json.Marshal(query)
	if err != nil {
		fmt.Printf("Error querying records by %s: %v\n", fieldName, err)
		return "[]", nil
	}

	fmt.Printf("Query: %s\n", string(queryString))

	resultsIterator, err := ctx.GetStub().GetQueryResult(string(queryString))
	if err != nil {
		fmt.Printf("Error querying records by %s: %v\n", fieldName, err)
		return "[]", nil
	}
	defer resultsIterator.Close()

	results, err := getAllResults(resultsIterator)
	if err != nil {
		fmt.Printf("Error querying records by %s: %v\n", fieldName, err)
		return "[]", nil
	}

	resultsJSON, err := json.Marshal(results)
	if err != nil {
		fmt.Printf("Error querying records by %s: %v\n", fieldName, err)
		return "[]", nil
	}

	fmt.Println("============= END : Query Records By Field ===========")
	return string(resultsJSON), nil
}

// QueryTicket queries a ticket by key
func (qu *QueryUtils) QueryTicket(ctx contractapi.TransactionContextInterface, ticketKey string) (string, error) {
	ticketAsBytes, err := ctx.GetStub().GetState(ticketKey)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}
	if ticketAsBytes == nil || len(ticketAsBytes) == 0 {
		return "", fmt.Errorf("Ticket %s does not exist", ticketKey)
	}
	return string(ticketAsBytes), nil
}

// QueryAllTickets queries all tickets
func (qu *QueryUtils) QueryAllTickets(ctx contractapi.TransactionContextInterface) (string, error) {
	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"docType": "ticket",
		},
	}

	queryString, err := json.Marshal(query)
	if err != nil {
		return "", err
	}

	resultsIterator, err := ctx.GetStub().GetQueryResult(string(queryString))
	if err != nil {
		return "", err
	}
	defer resultsIterator.Close()

	allResults, err := getAllResults(resultsIterator)
	if err != nil {
		return "", err
	}

	resultsJSON, err := json.Marshal(allResults)
	if err != nil {
		return "", err
	}

	return string(resultsJSON), nil
}

// QueryTicketsByField queries tickets by a specific field
func (qu *QueryUtils) QueryTicketsByField(ctx contractapi.TransactionContextInterface, fieldName string, fieldValue string) (string, error) {
	fmt.Printf("============= START : Query Tickets By Field %s ===========\n", fieldName)

	var parsedValue interface{} = fieldValue

	if fieldName == "incrementId" || fieldName == "receivedWeight" {
		floatVal, err := strconv.ParseFloat(fieldValue, 64)
		if err != nil {
			fmt.Printf("Error querying tickets by %s: Invalid %s: %s is not a number\n", fieldName, fieldName, fieldValue)
			return "[]", nil
		}
		parsedValue = floatVal
	}

	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"docType": "ticket",
			fieldName: parsedValue,
		},
	}

	queryString, err := json.Marshal(query)
	if err != nil {
		fmt.Printf("Error querying tickets by %s: %v\n", fieldName, err)
		return "[]", nil
	}

	fmt.Printf("Query: %s\n", string(queryString))

	resultsIterator, err := ctx.GetStub().GetQueryResult(string(queryString))
	if err != nil {
		fmt.Printf("Error querying tickets by %s: %v\n", fieldName, err)
		return "[]", nil
	}
	defer resultsIterator.Close()

	results, err := getAllResults(resultsIterator)
	if err != nil {
		fmt.Printf("Error querying tickets by %s: %v\n", fieldName, err)
		return "[]", nil
	}

	resultsJSON, err := json.Marshal(results)
	if err != nil {
		fmt.Printf("Error querying tickets by %s: %v\n", fieldName, err)
		return "[]", nil
	}

	fmt.Println("============= END : Query Tickets By Field ===========")
	return string(resultsJSON), nil
}

// GetRecordHistory gets the history of a record
func (qu *QueryUtils) GetRecordHistory(ctx contractapi.TransactionContextInterface, recordId string) (string, error) {
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(recordId)
	if err != nil {
		return "", err
	}
	defer resultsIterator.Close()

	results, err := getAllResults(resultsIterator)
	if err != nil {
		return "", err
	}

	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return "", err
	}

	return string(resultsJSON), nil
}
