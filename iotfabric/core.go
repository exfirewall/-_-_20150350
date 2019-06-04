package iotfabric

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"time"
)

type Device struct {
	Id          string
	humidity    float64
	temperature float64
	Timestamp   time.Time
}

type IotFabric interface {
	AddDevice(shim.ChaincodeStubInterface, *Device) error
	CheckDevice(shim.ChaincodeStubInterface, string) (bool, error)
	ValidateDevice(shim.ChaincodeStubInterface, *Device) (bool, error)
	GetDevice(shim.ChaincodeStubInterface, string) (*Device, error)
	UpdateDevice(shim.ChaincodeStubInterface, *Device) error
	ListDevices(shim.ChaincodeStubInterface) ([]*Device, error)

	TransferData(stub shim.ChaincodeStubInterface, DeviceId string, newDeviceId string) error
}
