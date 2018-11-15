package service

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/chclient"
)

// 保存票据
// return string: 为链码层执行后返回的内容;  error: 当前方法执行后返回的错误信息
func (setup *FabricSetupService) SaveBill(bill Bill) (string, error) {

	var args []string
	args = append(args, "issue")

	// 将票据对象序列化成为字节数组
	b, err := json.Marshal(bill)
	if err != nil {
		return "", fmt.Errorf("指定的票据对象序列化时发生错误")
	}

	// 设置调用链码执行交易的请求参数
	req := chclient.Request{ChaincodeID: setup.Fabric.ChaincodeID, Fcn: args[0], Args: [][]byte{b}}

	// 调用链码执行交易
	response, err := setup.Fabric.Client.Execute(req)
	if err != nil {
		return "", fmt.Errorf("发布新票据失败: %v", err)
	}

	return response.TransactionID.ID, nil

}

// 查询持有人的票据列表
// args: holdrCmID
func (setup *FabricSetupService) FindBills(holdrCmID string) ([]byte, error) {

	var args []string
	args = append(args, "queryBills")
	args = append(args, holdrCmID)

	// 设置请求参数
	req := chclient.Request{ChaincodeID: setup.Fabric.ChaincodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1])}}

	// 调用链码
	response, err := setup.Fabric.Client.Query(req)
	if err != nil {
		return []byte{0x00}, fmt.Errorf(err.Error())
	}

	b := response.Payload

	return b[:], nil
}

// 发起背书请求
// args: billNo, waitEndorseCmID waitEndorseAcct
func (setup *FabricSetupService) Endorse(billNo string, endorseCmID string, endorseAcct string) (string, error) {
	var args []string
	args = append(args, "endorse")
	args = append(args, billNo)
	args = append(args, endorseCmID)
	args = append(args, endorseAcct)

	req := chclient.Request{ChaincodeID: setup.Fabric.ChaincodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1]), []byte(args[2]), []byte(args[3])}}

	response, err := setup.Fabric.Client.Execute(req)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	//response.TransactionID.ID

	return string(response.Payload), nil
}

// 查询待背书票据列表
// args: waitEndorseCmID
func (setup *FabricSetupService) FindWaitBills(waitEndorseCmID string) ([]byte, error) {
	var args []string
	args = append(args, "queryWaitBills")
	args = append(args, waitEndorseCmID)

	req := chclient.Request{ChaincodeID: setup.Fabric.ChaincodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1])}}

	response, err := setup.Fabric.Client.Query(req)
	if err != nil {
		return []byte{0x00}, fmt.Errorf(err.Error())
	}

	b := response.Payload
	return b[:], nil
}

// 签收票据
// args: billNo, waitEndorseID, waitEndorseAcct
func (setup *FabricSetupService) Accept(billNo string, endorseCmID string, endorseAcct string) (string, error) {
	var args []string
	args = append(args, "accept")
	args = append(args, billNo)
	args = append(args, endorseCmID)
	args = append(args, endorseAcct)

	req := chclient.Request{ChaincodeID: setup.Fabric.ChaincodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1]), []byte(args[2]), []byte(args[3])}}
	respone, err := setup.Fabric.Client.Execute(req)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	return string(respone.Payload), nil

}

// 根据票据号码查询票据详情
// args: billNo
func (setup *FabricSetupService) FindBillByNo(billNo string) ([]byte, error) {
	var args []string
	args = append(args, "queryBillByNo")
	args = append(args, billNo)

	req := chclient.Request{ChaincodeID: setup.Fabric.ChaincodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1])}}

	response, err := setup.Fabric.Client.Query(req)
	if err != nil {
		return []byte{0x00}, fmt.Errorf("查询指定的票据详情失败: " + err.Error())
	}
	return response.Payload, nil
}

// 票据拒签
// args: billNo, waitEndorseID, waitEndorseAcct
func (setup *FabricSetupService) Reject(billNo string, endorseCmID string, endorseAcct string) (string, error) {
	var args []string
	args = append(args, "reject")
	args = append(args, billNo)
	args = append(args, endorseCmID)
	args = append(args, endorseAcct)

	req := chclient.Request{ChaincodeID: setup.Fabric.ChaincodeID, Fcn: args[0], Args: [][]byte{[]byte(args[1]), []byte(args[2]), []byte(args[3])}}

	response, err := setup.Fabric.Client.Execute(req)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	return string(response.Payload), nil
}
