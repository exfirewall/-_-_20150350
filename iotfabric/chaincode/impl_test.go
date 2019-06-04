package chaincode_test

import (
	"github.com/hyperledger/fabric/chaincode/shim"
	"github.com/hyperledger/fabric/common/util"
	"github.com/stretchr/testify/assert"
	"iotfabric/chaincode"
	"testing"
)

const (
	DeviceOneJSON = `{"Id":"1", "humidity":"50", "temperature":"20"}`

	DevicesJSON = "[" + DeviceOneJSON + "]"
)

func TestInit(t *testing.T) {
	stub := shim.NewMockStub("iotfabric", new(chaincode.IotCC))
	if assert.NotNil(t, stub) {
		res := stub.MockInit(util.GenerateUUID(), nil)
		assert.True(t, res.Status < shim.ERRORTHRESHOLD)
	}
}

func TestInvoke(t *testing.T) {
	stub := shim.NewMockStub("iotfabric", new(chaincide.IotCC))
	if !assert.NotNil(t, stub) {
		return
	}

	if !assert.True(t, stub.MockInit(util.GenerateUUID(), nil).Status < shim.ERRORTHRESHOLD) {
		return
	}

	if !assert.True(
		t,
		stub.MockInvoke(
			util.GenerateUUID(),
			getBytes("AddDevice", DeviceOneJSON),
		).Status < shim.ERRORTHRESHOLD,
	) {
		return
	}

	res := stub.MockInvoke(util.GenerateUUID(), getBytes("ListDevices"))
	_ = assert.True(t, res.Status < shim.ERRORTHRESHOLD) &&
		assert.JSONEq(t, DevicesJSON, string(res.Payload))
}

func getBytes(function string, args ...string) [][]byte {
	bytes := make([][]byte, 0, len(args)+1)
	bytes = append(bytes, []byte(function))
	for _, s := range args {
		bytes = append(bytes, []byte(s))
	}

	return bytes
}
