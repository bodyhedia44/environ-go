package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ProofRecord represents a proof record
type ProofRecord struct {
	SponsorID       string   `json:"sponsor_id"`
	ProofShortID    string   `json:"proof_short_id"`
	CollectorName   string   `json:"collector_name"`
	BulkName        string   `json:"bulk_name"`
	ParentIncrement float64  `json:"parent_increment"`
	ChainedWeight   float64  `json:"chained_weight"`
	TraceChainType  string   `json:"traceChainType"`
	BulkShortID     string   `json:"bulk_short_id"`
	StoreIncrement  *float64 `json:"store_increment"`
	PressIncrement  *float64 `json:"press_increment"`
	RecordID        string   `json:"recordId"`
	CreatedAt       string   `json:"createdAt"`
	CreatedBy       string   `json:"createdBy"`
	DocType         string   `json:"docType"`
}

// ProofRecordManager handles proof record operations
type ProofRecordManager struct{}

// NewProofRecordManager creates a new ProofRecordManager instance
func NewProofRecordManager() *ProofRecordManager {
	return &ProofRecordManager{}
}

// CreateProofRecordResponse represents the response from creating a proof record
type CreateProofRecordResponse struct {
	Success         bool                     `json:"success"`
	Message         string                   `json:"message"`
	RecordID        string                   `json:"recordId,omitempty"`
	Record          *ProofRecord             `json:"record,omitempty"`
	DuplicateFields []string                 `json:"duplicateFields,omitempty"`
	ExistingRecords []map[string]interface{} `json:"existingRecords,omitempty"`
}

// CreateProofRecord creates a new proof record
func (prm *ProofRecordManager) CreateProofRecord(ctx contractapi.TransactionContextInterface, recordData string) (string, error) {
	fmt.Println("============= START : Create Proof Record ===========")

	var record map[string]interface{}
	err := json.Unmarshal([]byte(recordData), &record)
	if err != nil {
		response := CreateProofRecordResponse{
			Success: false,
			Message: fmt.Sprintf("Error creating proof record: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	if !prm.validateRecord(record) {
		response := CreateProofRecordResponse{
			Success: false,
			Message: "Error creating proof record: Invalid record data. Missing required fields.",
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	duplicateCheckResult, err := prm.checkForDuplicates(ctx, record)
	if err != nil {
		response := CreateProofRecordResponse{
			Success: false,
			Message: fmt.Sprintf("Error creating proof record: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	if duplicateCheckResult.IsDuplicate {
		response := CreateProofRecordResponse{
			Success:         false,
			Message:         "Duplicate record found",
			DuplicateFields: duplicateCheckResult.DuplicateFields,
			ExistingRecords: duplicateCheckResult.ExistingRecords,
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	recordKey := GenerateRecordKey(record)
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		clientID = ""
	}

	record["recordId"] = recordKey
	record["createdAt"] = time.Now().UTC().Format(time.RFC3339)
	record["createdBy"] = clientID
	record["docType"] = "proofRecord"

	recordJSON, err := json.Marshal(record)
	if err != nil {
		response := CreateProofRecordResponse{
			Success: false,
			Message: fmt.Sprintf("Error creating proof record: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	err = ctx.GetStub().PutState(recordKey, recordJSON)
	if err != nil {
		response := CreateProofRecordResponse{
			Success: false,
			Message: fmt.Sprintf("Error creating proof record: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	fmt.Println("============= END : Create Proof Record ===========")

	var proofRecord ProofRecord
	json.Unmarshal(recordJSON, &proofRecord)

	response := CreateProofRecordResponse{
		Success:  true,
		Message:  "Record saved successfully",
		RecordID: recordKey,
		Record:   &proofRecord,
	}
	responseJSON, _ := json.Marshal(response)
	return string(responseJSON), nil
}

func (prm *ProofRecordManager) validateRecord(record map[string]interface{}) bool {
	requiredFields := []string{
		"sponsor_id",
		"proof_short_id",
		"collector_name",
		"bulk_name",
		"parent_increment",
		"chained_weight",
		"traceChainType",
		"bulk_short_id",
	}

	for _, field := range requiredFields {
		value, exists := record[field]
		if !exists || value == nil {
			return false
		}
	}
	return true
}

// DuplicateCheckResult represents the result of a duplicate check
type DuplicateCheckResult struct {
	IsDuplicate     bool                     `json:"isDuplicate"`
	DuplicateFields []string                 `json:"duplicateFields,omitempty"`
	ExistingRecords []map[string]interface{} `json:"existingRecords,omitempty"`
}

func (prm *ProofRecordManager) checkForDuplicates(ctx contractapi.TransactionContextInterface, record map[string]interface{}) (*DuplicateCheckResult, error) {
	storeIncrement, hasStore := record["store_increment"]
	pressIncrement, hasPress := record["press_increment"]

	storeIsNull := !hasStore || storeIncrement == nil
	pressIsNull := !hasPress || pressIncrement == nil

	conditions := []struct {
		condition   bool
		checkFields []string
	}{
		{
			condition:   !storeIsNull,
			checkFields: []string{"store_increment", "press_increment", "parent_increment"},
		},
		{
			condition:   storeIsNull,
			checkFields: []string{"press_increment", "parent_increment"},
		},
		{
			condition:   pressIsNull,
			checkFields: []string{"parent_increment", "store_increment"},
		},
		{
			condition:   pressIsNull && storeIsNull,
			checkFields: []string{"parent_increment"},
		},
	}

	for _, cond := range conditions {
		if cond.condition {
			duplicateResult, err := prm.findDuplicatesByFields(ctx, record, cond.checkFields)
			if err != nil {
				return nil, err
			}
			if duplicateResult.IsDuplicate {
				return &DuplicateCheckResult{
					IsDuplicate:     true,
					DuplicateFields: cond.checkFields,
					ExistingRecords: duplicateResult.ExistingRecords,
				}, nil
			}
		}
	}

	return &DuplicateCheckResult{IsDuplicate: false}, nil
}

func (prm *ProofRecordManager) findDuplicatesByFields(ctx contractapi.TransactionContextInterface, record map[string]interface{}, fields []string) (*DuplicateCheckResult, error) {
	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"docType": "proofRecord",
		},
	}

	selector := query["selector"].(map[string]interface{})
	for _, field := range fields {
		if value, exists := record[field]; exists && value != nil {
			selector[field] = value
		}
	}

	queryString, err := json.Marshal(query)
	if err != nil {
		fmt.Printf("Error finding duplicates: %v\n", err)
		return &DuplicateCheckResult{IsDuplicate: false, ExistingRecords: []map[string]interface{}{}}, nil
	}

	resultsIterator, err := ctx.GetStub().GetQueryResult(string(queryString))
	if err != nil {
		fmt.Printf("Error finding duplicates: %v\n", err)
		return &DuplicateCheckResult{IsDuplicate: false, ExistingRecords: []map[string]interface{}{}}, nil
	}
	defer resultsIterator.Close()

	existingRecords, err := getAllResults(resultsIterator)
	if err != nil {
		fmt.Printf("Error finding duplicates: %v\n", err)
		return &DuplicateCheckResult{IsDuplicate: false, ExistingRecords: []map[string]interface{}{}}, nil
	}

	return &DuplicateCheckResult{
		IsDuplicate:     len(existingRecords) > 0,
		ExistingRecords: existingRecords,
	}, nil
}
