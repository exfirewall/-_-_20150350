package main

import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"iotfabric/chaincode"
)

func main() {
	err := shim.Start(new(chaincode.IotCC))
	if err != nil {
		fmt.Printf("Error in chaincode process: %s\n", err)
	}
}
