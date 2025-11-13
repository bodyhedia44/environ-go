package main

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ProofRecordsContract provides functions for managing proof records and tickets
type ProofRecordsContract struct {
	contractapi.Contract
}

// InitLedger initializes the ledger
func (c *ProofRecordsContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("============= START : Initialize Ledger ===========")
	fmt.Println("============= END : Initialize Ledger ===========")
	return nil
}

// CreateProofRecord creates a new proof record
func (c *ProofRecordsContract) CreateProofRecord(ctx contractapi.TransactionContextInterface, recordData string) (string, error) {
	manager := NewProofRecordManager()
	return manager.CreateProofRecord(ctx, recordData)
}

// QueryProofRecord queries a proof record by ID
func (c *ProofRecordsContract) QueryProofRecord(ctx contractapi.TransactionContextInterface, recordId string) (string, error) {
	queryUtils := NewQueryUtils()
	return queryUtils.QueryProofRecord(ctx, recordId)
}

// QueryAllProofRecords queries all proof records
func (c *ProofRecordsContract) QueryAllProofRecords(ctx contractapi.TransactionContextInterface) (string, error) {
	queryUtils := NewQueryUtils()
	return queryUtils.QueryAllProofRecords(ctx)
}

// QueryRecordsByField queries records by a specific field
func (c *ProofRecordsContract) QueryRecordsByField(ctx contractapi.TransactionContextInterface, fieldName string, fieldValue string) (string, error) {
	queryUtils := NewQueryUtils()
	return queryUtils.QueryRecordsByField(ctx, fieldName, fieldValue)
}

// GetRecordHistory gets the history of a record
func (c *ProofRecordsContract) GetRecordHistory(ctx contractapi.TransactionContextInterface, recordId string) (string, error) {
	queryUtils := NewQueryUtils()
	return queryUtils.GetRecordHistory(ctx, recordId)
}

// CreateTicket creates a new ticket
func (c *ProofRecordsContract) CreateTicket(ctx contractapi.TransactionContextInterface, ticketData string) (string, error) {
	manager := NewTicketManager()
	return manager.CreateTicket(ctx, ticketData)
}

// QueryTicket queries a ticket by key
func (c *ProofRecordsContract) QueryTicket(ctx contractapi.TransactionContextInterface, ticketKey string) (string, error) {
	queryUtils := NewQueryUtils()
	return queryUtils.QueryTicket(ctx, ticketKey)
}

// QueryAllTickets queries all tickets
func (c *ProofRecordsContract) QueryAllTickets(ctx contractapi.TransactionContextInterface) (string, error) {
	queryUtils := NewQueryUtils()
	return queryUtils.QueryAllTickets(ctx)
}

// QueryTicketsByField queries tickets by a specific field
func (c *ProofRecordsContract) QueryTicketsByField(ctx contractapi.TransactionContextInterface, fieldName string, fieldValue string) (string, error) {
	queryUtils := NewQueryUtils()
	return queryUtils.QueryTicketsByField(ctx, fieldName, fieldValue)
}

// CompareWeightsByPressIncrement compares weights by press increment
func (c *ProofRecordsContract) CompareWeightsByPressIncrement(ctx contractapi.TransactionContextInterface, deleteViolations string) (string, error) {
	weightComp := NewWeightComparison()
	return weightComp.CompareWeightsByPressIncrement(ctx, deleteViolations)
}

// CompareWeightsByStoreIncrement compares weights by store increment
func (c *ProofRecordsContract) CompareWeightsByStoreIncrement(ctx contractapi.TransactionContextInterface, deleteViolations string) (string, error) {
	weightComp := NewWeightComparison()
	return weightComp.CompareWeightsByStoreIncrement(ctx, deleteViolations)
}
