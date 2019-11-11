package core

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ElrondNetwork/elrond-go/process/smartContract"
	vmcommon "github.com/ElrondNetwork/elrond-vm-common"
	"github.com/gin-gonic/gin"
)

// VMValueRequest represents the structure on which user input for generating a new transaction will validate against
type VMValueRequest struct {
	OnTestnet           bool     `form:"onTestnet" json:"onTestnet"`
	TestnetNodeEndpoint string   `form:"testnetNodeEndpoint" json:"testnetNodeEndpoint"`
	ScAddress           string   `form:"scAddress" json:"scAddress"`
	FuncName            string   `form:"funcName" json:"funcName"`
	Args                []string `form:"args"  json:"args"`
}

func handlerGetHex(context *gin.Context) {
	doGetVMValue(context, vmcommon.AsHex)
}

func handlerGetString(context *gin.Context) {
	doGetVMValue(context, vmcommon.AsString)
}

func handlerGetInt(context *gin.Context) {
	doGetVMValue(context, vmcommon.AsBigIntString)
}

func doGetVMValue(context *gin.Context, asType vmcommon.ReturnDataKind) {
	vmOutput, err := doExecuteQuery(context)

	if err != nil {
		returnBadRequest(context, "doGetVMValue", err)
		return
	}

	returnData, err := vmOutput.GetFirstReturnData(asType)
	if err != nil {
		returnBadRequest(context, "doGetVMValue", err)
		return
	}

	returnOkResponse(context, returnData)
}

func handlerExecuteQuery(context *gin.Context) {
	vmOutput, err := doExecuteQuery(context)
	if err != nil {
		returnBadRequest(context, "executeQuery", err)
		return
	}

	returnOkResponse(context, vmOutput)
}

func doExecuteQuery(context *gin.Context) (*vmcommon.VMOutput, error) {
	facade, _ := context.MustGet("elrondFacade").(FacadeHandler)

	request := VMValueRequest{}
	err := context.ShouldBindJSON(&request)
	if err != nil {
		return nil, err
	}

	command, err := createSCQuery(&request)
	if err != nil {
		return nil, err
	}

	vmOutput, err := facade.ExecuteSCQuery(command)
	if err != nil {
		return nil, err
	}

	return vmOutput, nil
}

func createSCQuery(request *VMValueRequest) (*smartContract.SCQuery, error) {
	decodedAddress, err := hex.DecodeString(request.ScAddress)
	if err != nil {
		return nil, fmt.Errorf("'%s' is not a valid hex string: %s", request.ScAddress, err.Error())
	}

	argumentsAsInt := make([]*big.Int, 0)
	for _, arg := range request.Args {
		argBytes, err := hex.DecodeString(arg)
		if err != nil {
			return nil, fmt.Errorf("'%s' is not a valid hex string: %s", arg, err.Error())
		}

		argumentsAsInt = append(argumentsAsInt, big.NewInt(0).SetBytes(argBytes))
	}

	return &smartContract.SCQuery{
		ScAddress: decodedAddress,
		FuncName:  request.FuncName,
		Arguments: argumentsAsInt,
	}, nil
}