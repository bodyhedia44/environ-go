package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	proofRecordsContract := new(ProofRecordsContract)

	chaincode, err := contractapi.NewChaincode(proofRecordsContract)
	if err != nil {
		log.Panicf("Error creating proof records chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting proof records chaincode: %v", err)
	}
}
