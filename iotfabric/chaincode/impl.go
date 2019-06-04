package chaincode

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"github.com/jinzhu/inflection"
	"iotfabric"
)

func checkLen(logger *shim.ChaincodeLogger, expected int, args []string) error {
	if len(args) < expected {
		mes := fmt.Sprintf(
			"not enough number of arguments: %d given, %d expected",
			len(args),
			expected,
		)
		logger.Warning(mes)
		return errors.New(mes)
	}
	return nil
}

type IotCC struct {
}

func (this *IotCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger := shim.NewLogger("iotfabric")
	logger.Info("chaincode initialized")
	return shim.Success([]byte{})
}
func (this *IotCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger := shim.NewLogger("iotfabric")

	timestamp, err := stub.GetTxTimestamp()
	if err != nil {
		return shim.Error(fmt.Sprintf("failed to get TX timestamp: %s", err))
	}
	logger.Infof(
		"Invoke called: Tx ID = %s, timestamp = %s",
		stub.GetTxID(),
		timestamp,
	)

	var (
		fcn  string
		args []string
	)
	fcn, args = stub.GetFunctionAndParameters()
	logger.Infof("function name = %s", fcn)

	switch fcn {

	case "AddDevice":
		if err := checkLen(logger, 1, args); err != nil {
			return shim.Error(err.Error())
		}
		device := new(iotfabric.Device)
		err := json.Unmarshal([]byte(args[0]), device)
		if err != nil {
			mes := fmt.Sprintf("failed to unmarshal Device JSON: %s", err.Error())
			logger.Warning(mes)
			return shim.Error(mes)
		}

		err = this.AddDevice(stub, device)
		if err != nil {
			return shim.Error(err.Error())
		}

		return shim.Success([]byte{})
	case "ListDevice":
		devices, err := this.ListDevice(stub)
		if err != nil {
			return shim.Error(err.Error())
		}

		//marshal
		b, err := json.Marshal(devices)
		if err != nil {
			mes := fmt.Sprintf("failed to marshal Devices: %s", err.Error())
			logger.Warning(mes)
			return shim.Error(mes)
		}

		return shim.Success(b)
	case "UpdateDevice":
		if err := checkLen(logger, 1, args); err != nil {
			return shim.Error(err.Error())
		}

		//unmarshal
		device := new(iotfabric.Device)
		err := json.Unmarshal([]byte(args[0]), device)
		if err != nil {
			mes := fmt.Sprintf("failed to unmarshal Device JSON: %s", err.Error())
			logger.Warning(mes)
			return shim.Error(mes)
		}

		err = this.UpdateDevice(stub, device)
		if err != nil {
			return shim.Error(err.Error())
		}

		return shim.Success([]byte{})
	case "TransferData":
		if err := checkLen(logger, 2, args); err != nil {
			return shim.Error(err.Error())
		}

		//unmarshal
		var Hum, Temp string
		var newDeviceId string
		err := json.Unmarshal([]byte(args[0]), &Hum)
		if err != nil {
			mes := fmt.Sprintf(
				"failed to unmarshal the 1st argument: %s",
				err.Error(),
			)
			logger.Warning(mes)
			return shim.Error(mes)
		}
		err = json.Unmarshal([]byte(args[1]), &Temp)
		if err != nil {
			mes := fmt.Sprintf(
				"failed to unmarshal the 2nd argument: %s",
				err.Error(),
			)
			logger.Warning(mes)
			return shim.Error(mes)
		}
		err = json.Unmarshal([]byte(args[2]), &newDeviceId)
		if err != nil {
			mes := fmt.Sprintf(
				"failed to unmarshal the 3rd argument: %s",
				err.Error(),
			)
			logger.Warning(mes)
			return shim.Error(mes)
		}
		err = this.TransferData(stub, Hum, newDeviceId)
		if err != nil {
			return shim.Error(err.Error())
		}

		return shim.Success([]byte{})
	}
	mes := fmt.Sprintf("Unknown method: %s", fcn)
	logger.Warning(mes)
	return shim.Error(mes)
}

func (this *IotCC) AddDevice(stub shim.ChaincodeStubInterface, device *iotfabric.Device) error {
	logger := shim.NewLogger("iotfabric")
	logger.Infof("AddDevice: Id = %s", device.Id)

	key, err := stub.CreateCompositeKey("Device", []string{device.Id})
	if err != nil {
		logger.Warning(err.Error())
		return err
	}
	found, err := this.CheckDevice(stub, device.Id)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}
	if found {
		mes := fmt.Sprintf("Device with Id = %s already exists", device.Id)
		logger.Warning(mes)
		return errors.New(mes)
	}

	ok, err := this.ValidateDevice(stub, device)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}
	if !ok {
		mes := "Validation of the Device failed"
		logger.Warning(mes)
		return errors.New(mes)
	}
	//converts to JSON
	b, err := json.Marshal(device)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}
	//sotres to the State DB
	err = stub.PutState(key, b)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}

	return nil
}

func (this *IotCC) CheckDevice(stub shim.ChaincodeStubInterface, id string) (bool, error) {
	logger := shim.NewLogger("iotfabric")
	logger.Infof("CheckDevice: Id = %s", id)
	//creates a composite key
	key, err := stub.CreateCompositeKey("Device", []string{id})
	if err != nil {
		logger.Warning(err.Error())
		return false, err
	}

	//loads from the State DB
	jsonBytes, err := stub.GetState(key)
	if err != nil {
		logger.Warning(err.Error())
		return false, err
	}

	return jsonBytes != nil, nil
}

func (this *IotCC) ValidateDevice(stub shim.ChaincodeStubInterface, device *iotfabric.Device) (bool, error) {
	logger := shim.NewLogger("iotfabric")
	logger.Infof("ValidateDevice: Id = %s", device.Id)

	found, err := this.CheckDevice(stub, device.Id)
	if err != nil {
		logger.Warning(err.Error())
		return false, err
	}

	return found, nil
}
func (this *IotCC) GetDevice(stub shim.ChaincodeStubInterface, id string) (*iotfabric.Device, error) {
	logger := shim.NewLogger("iotfabric")
	logger.Infof("GetDevice: Id = %s", id)

	key, err := stub.CreateCompositeKey("Device", []string{id})
	if err != nil {
		logger.Warning(err.Error())
		return nil, err
	}

	jsonBytes, err := stub.GetState(key)
	if err != nil {
		logger.Warning(err.Error())
		return nil, err
	}

	if jsonBytes == nil {
		mes := fmt.Sprintf("Device whith Id = %s was not found", id)
		logger.Warning(mes)
		return nil, errors.New(mes)
	}

	device := new(iotfabric.Device)
	err = json.Unmarshal(jsonBytes, device)
	if err != nil {
		logger.Warning(err.Error())
		return nil, err
	}

	return device, nil

}

func (this *IotCC) UpdateDevice(stub shim.ChaincodeStubInterface, device *iotfabric.Device) error {
	logger := shim.NewLogger("iotfabric")
	logger.Infof("UpdateDevice: device = %+V", device)

	found, err := this.CheckDevice(stub, device.Id)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}
	if !found {
		mes := fmt.Sprintf("Device wiht Id = %s does not exist", device.Id)
		logger.Warning(mes)
		return errors.New(mes)
	}

	ok, err := this.ValidateDevice(stub, device)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}
	if !ok {
		mes := "Validation of the Device failed"
		logger.Warning(mes)
		return errors.New(mes)
	}
	key, err := stub.CreateCompositeKey("Device", []string{device.Id})
	if err != nil {
		logger.Warning(err.Error())
		return err
	}
	b, err := json.Marshal(device)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}

	err = stub.PutState(key, b)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}

	return nil
}

func (this *IotCC) ListDevice(stub shim.ChaincodeStubInterface) ([]*iotfabric.Device, error) {
	logger := shim.NewLogger("iotfabric")
	logger.Info("ListDevice")

	iter, err := stub.GetStateByPartialCompositeKey("Device", []string{})
	if err != nil {
		logger.Warning(err.Error())
		return nil, err
	}

	defer iter.Close()

	devices := []*iotfabric.Device{}
	for iter.HasNext() {
		kv, err := iter.Next()
		if err != nil {
			logger.Warning(err.Error())
			return nil, err
		}
		device := new(iotfabric.Device)
		err = json.Unmarshal(kv.Value, device)
		if err != nil {
			logger.Warning(err.Error())
			return nil, err
		}
		devices = append(devices, device)
	}

	if len(devices) > 1 {
		logger.Infof("%d %s found", len(devices), inflection.Plural("Device"))
	} else {
		logger.Infof("%d %s found", len(devices), "Car")
	}

	return devices, nil
}

func (this *IotCC) TransferData(stub shim.ChaincodeStubInterface, deviceId string, newDeviceId string) error {
	logger := shim.NewLogger("iotfabric")
	logger.Infof("TransferData: Device Id = %s, new Device Id = %s", deviceId, newDeviceId)

	device, err := this.GetDevice(stub, deviceId)
	if err != nil {
		logger.Warning(err.Error())
		return err
	}
	device.Id = newDeviceId

	err = this.UpdateDevice(stub, device)

	if err != nil {
		logger.Warning(err.Error())
		return err
	}

	return nil
}
