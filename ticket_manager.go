package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Ticket represents a ticket record
type Ticket struct {
	ID             string  `json:"id"`
	ReceivedWeight float64 `json:"receivedWeight"`
	IncrementID    float64 `json:"incrementId"`
	CreatedAt      string  `json:"createdAt"`
	CreatedBy      string  `json:"createdBy"`
	DocType        string  `json:"docType"`
}

// TicketManager handles ticket operations
type TicketManager struct{}

// NewTicketManager creates a new TicketManager instance
func NewTicketManager() *TicketManager {
	return &TicketManager{}
}

// CreateTicketResponse represents the response from creating a ticket
type CreateTicketResponse struct {
	Success         bool                     `json:"success"`
	Message         string                   `json:"message"`
	TicketKey       string                   `json:"ticketKey,omitempty"`
	Ticket          *Ticket                  `json:"ticket,omitempty"`
	DuplicateFields []string                 `json:"duplicateFields,omitempty"`
	ExistingTickets []map[string]interface{} `json:"existingTickets,omitempty"`
}

// CreateTicket creates a new ticket
func (tm *TicketManager) CreateTicket(ctx contractapi.TransactionContextInterface, ticketData string) (string, error) {
	fmt.Println("============= START : Create Ticket ===========")

	var ticket map[string]interface{}
	err := json.Unmarshal([]byte(ticketData), &ticket)
	if err != nil {
		response := CreateTicketResponse{
			Success: false,
			Message: fmt.Sprintf("Error creating ticket: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	if !tm.validateTicket(ticket) {
		response := CreateTicketResponse{
			Success: false,
			Message: "Error creating ticket: Invalid ticket data. Missing required fields.",
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	duplicateCheckResult, err := tm.checkTicketForDuplicates(ctx, ticket)
	if err != nil {
		response := CreateTicketResponse{
			Success: false,
			Message: fmt.Sprintf("Error creating ticket: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	if duplicateCheckResult.IsDuplicate {
		response := CreateTicketResponse{
			Success:         false,
			Message:         "Duplicate ticket found",
			DuplicateFields: duplicateCheckResult.DuplicateFields,
			ExistingTickets: duplicateCheckResult.ExistingRecords,
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	ticketID := ticket["id"].(string)
	ticketKey := fmt.Sprintf("TICKET_%s", ticketID)

	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		clientID = ""
	}

	ticket["createdAt"] = time.Now().UTC().Format(time.RFC3339)
	ticket["createdBy"] = clientID
	ticket["docType"] = "ticket"

	ticketJSON, err := json.Marshal(ticket)
	if err != nil {
		response := CreateTicketResponse{
			Success: false,
			Message: fmt.Sprintf("Error creating ticket: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	err = ctx.GetStub().PutState(ticketKey, ticketJSON)
	if err != nil {
		response := CreateTicketResponse{
			Success: false,
			Message: fmt.Sprintf("Error creating ticket: %v", err),
		}
		responseJSON, _ := json.Marshal(response)
		return string(responseJSON), nil
	}

	fmt.Println("============= END : Create Ticket ===========")

	var ticketRecord Ticket
	json.Unmarshal(ticketJSON, &ticketRecord)

	response := CreateTicketResponse{
		Success:   true,
		Message:   "Ticket saved successfully",
		TicketKey: ticketKey,
		Ticket:    &ticketRecord,
	}
	responseJSON, _ := json.Marshal(response)
	return string(responseJSON), nil
}

func (tm *TicketManager) validateTicket(ticket map[string]interface{}) bool {
	requiredFields := []string{
		"id",
		"receivedWeight",
		"incrementId",
	}

	for _, field := range requiredFields {
		value, exists := ticket[field]
		if !exists || value == nil {
			return false
		}
	}
	return true
}

// TicketDuplicateCheckResult represents the result of a ticket duplicate check
type TicketDuplicateCheckResult struct {
	IsDuplicate     bool                     `json:"isDuplicate"`
	DuplicateFields []string                 `json:"duplicateFields,omitempty"`
	ExistingRecords []map[string]interface{} `json:"existingTickets,omitempty"`
}

func (tm *TicketManager) checkTicketForDuplicates(ctx contractapi.TransactionContextInterface, ticket map[string]interface{}) (*TicketDuplicateCheckResult, error) {
	fields := []string{"incrementId", "id"}
	duplicateResult, err := tm.findTicketDuplicatesByFields(ctx, ticket, fields)
	if err != nil {
		return nil, err
	}

	if duplicateResult.IsDuplicate {
		return &TicketDuplicateCheckResult{
			IsDuplicate:     true,
			DuplicateFields: fields,
			ExistingRecords: duplicateResult.ExistingRecords,
		}, nil
	}

	return &TicketDuplicateCheckResult{IsDuplicate: false}, nil
}

func (tm *TicketManager) findTicketDuplicatesByFields(ctx contractapi.TransactionContextInterface, ticket map[string]interface{}, fields []string) (*TicketDuplicateCheckResult, error) {
	query := map[string]interface{}{
		"selector": map[string]interface{}{
			"docType": "ticket",
		},
	}

	selector := query["selector"].(map[string]interface{})
	for _, field := range fields {
		if value, exists := ticket[field]; exists && value != nil {
			selector[field] = value
		}
	}

	queryString, err := json.Marshal(query)
	if err != nil {
		fmt.Printf("Error finding ticket duplicates: %v\n", err)
		return &TicketDuplicateCheckResult{IsDuplicate: false, ExistingRecords: []map[string]interface{}{}}, nil
	}

	resultsIterator, err := ctx.GetStub().GetQueryResult(string(queryString))
	if err != nil {
		fmt.Printf("Error finding ticket duplicates: %v\n", err)
		return &TicketDuplicateCheckResult{IsDuplicate: false, ExistingRecords: []map[string]interface{}{}}, nil
	}
	defer resultsIterator.Close()

	existingTickets, err := getAllResults(resultsIterator)
	if err != nil {
		fmt.Printf("Error finding ticket duplicates: %v\n", err)
		return &TicketDuplicateCheckResult{IsDuplicate: false, ExistingRecords: []map[string]interface{}{}}, nil
	}

	return &TicketDuplicateCheckResult{
		IsDuplicate:     len(existingTickets) > 0,
		ExistingRecords: existingTickets,
	}, nil
}
