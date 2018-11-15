package main

import (
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// 发起背书请求
// arg: billNo, waitEndorseCmID, waitEndorseAcct
func (t *BillChaincode) endorse(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 3 {
		msg, _ := GetMsgString(1, "票据背书请求失败, 必须且只能指定票据号码, 待背书人证件号码, 待背书人名称三个值")
		return shim.Error(msg)
	}

	// 根据票据号码查询对应的票据状态
	bill, bl := GetBill(stub, args[0])
	if !bl {
		return shim.Error("票据背书请求失败, 查询票据状态时发生错误")
	}

	// 待背书人不能是当前持票人
	if bill.HoldrCmID == args[1] {
		return shim.Error("票据背书请求失败, 待背书人不能是当前持票人")
	}

	// 历史持有人不能为当前待背书人
	iterator, err := stub.GetHistoryForKey(args[0])
	if err != nil {
		return shim.Error("票据背书请求失败, 查询历史信息时发生错误")
	}
	defer iterator.Close()

	var hisBill Bill
	// 迭代处理
	for iterator.HasNext() {
		hisData, err := iterator.Next()
		if err != nil {
			return shim.Error("票据背书请求失败, 获取具体历史流转信息时发生错误")
		}

		if hisData.Value == nil {
			continue
		}

		err = json.Unmarshal(hisData.Value, &hisBill)
		if err != nil {
			return shim.Error("反序列化历史流转信息时发生错误")
		}

		// 历史持有人不能为当前待背书人
		if hisBill.HoldrCmID == args[1] {
			return shim.Error("票据背书请求失败, 历史持有人不能为当前待背书人")
		}

	}

	// 更改票据状态, 票据待背书人信息, 删除拒绝背书人信息
	bill.State = BillInfo_State_EndorseWaitSign
	bill.WaitEndorseCmID = args[1]
	bill.WaitEndorseAcct = args[2]

	bill.RejectEndorseCmID = ""
	bill.RejectEndorseAcct = ""

	// 保存票据
	_, bl = PutBill(stub, bill)
	if !bl {
		return shim.Error("票据背书请求失败, 保存状态时发生错误s")
	}

	// 增加复合键, 以便于批量查询待背书人的票据列表(根据待背书人证件号码及票据号码)
	//stub.CreateCompositeKey(WaitEndorName, []string{args[1], args[0]})
	waitEndorseCmIDBillInfoIDIndexKey, err := stub.CreateCompositeKey(IndexName, []string{bill.WaitEndorseCmID, bill.BillInfoID})
	if err != nil {
		return shim.Error("票据背书请求失败, 创建复合键时发生错误")
	}

	err = stub.PutState(waitEndorseCmIDBillInfoIDIndexKey, []byte{0x00})
	if err != nil {
		return shim.Error("票据背书请求失败, 保存复合键时发生错误")
	}

	/*msg, _ := GetMsgByte(0, "票据背书请求成功")
	return shim.Success(msg)*/
	return shim.Success([]byte("票据背书请求成功"))
}

// 查询待背书票据列表
// args: waitEndorseCmID
func (t *BillChaincode) queryWaitBills(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("必须且只能指定待背书人的证件号码")
	}

	// 根据待背书人证件号码查询复合键
	iterator, err := stub.GetStateByPartialCompositeKey(IndexName, []string{args[0]})
	if err != nil {
		return shim.Error("查询待背书票据失败, 查询对应的复合键时发生错误")
	}
	defer iterator.Close()

	// 迭代处理
	var bills []Bill
	for iterator.HasNext() {
		kv, err := iterator.Next()
		if err != nil {
			return shim.Error("查询待背书票据失败,获取复合键时发生错误")
		}

		_, composites, err := stub.SplitCompositeKey(kv.Key)
		if err != nil {
			return shim.Error("查询待背书票据失败, 分割复合键时发生错误")
		}

		bill, bl := GetBill(stub, composites[1])
		if !bl {
			return shim.Error("查询待背书票据失败, 查询具体的待背书票据时发生错误")
		}

		if bill.State == BillInfo_State_EndorseWaitSign && bill.WaitEndorseCmID == args[0] {
			bills = append(bills, bill)
		}
	}

	// 序列化查询结果
	b, err := json.Marshal(bills)
	if err != nil {
		return shim.Error("查询待背书票据失败, 序列化查询结果时发生错误")
	}
	return shim.Success(b)

}

// 背书签收
// args: billNo, endorseCmID, endorseAcct
func (t *BillChaincode) accept(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 3 {
		return shim.Error("票据背书签收失败, 必须且只能指定票据号码, 签收人证件号码, 签收人名称")
	}

	// 查询待修改的票据对象
	bill, bl := GetBill(stub, args[0])
	if !bl {
		return shim.Error("票据背书签收失败, 根据票据号码查询票据信息时发生错误")
	}

	// 创建复合键(根据前手持票人证件号码及票据号码)
	holdrCmIDBillInfoIDIndexKey, err := stub.CreateCompositeKey(IndexName, []string{bill.HoldrCmID, bill.BillInfoID})
	if err != nil {
		return shim.Error("背书签收失败, 创建复合键时发生错误")
	}

	// 根据复合键的Key从账本中删除信息, 以便于前手持票人无法查询到该票据的信息
	err = stub.DelState(holdrCmIDBillInfoIDIndexKey)
	if err != nil {
		return shim.Error("背书签收失败, 删除复合键时发生错误")
	}

	// 更改票据信息: 票据状态, 当前持票人信息, 待背书人信息
	bill.State = BillInfo_State_EndorseSigned
	bill.HoldrCmID = args[1]
	bill.HoldrAcct = args[2]

	bill.WaitEndorseCmID = ""
	bill.WaitEndorseAcct = ""

	_, bl = PutBill(stub, bill)
	if !bl {
		return shim.Error("飘扬背书签收失败, 保存票据信息时发生错误")
	}

	return shim.Success([]byte("票据背书签收成功"))
}

// 拒绝背书
// args: billNo, rejectCmID, rejectAcct
func (t *BillChaincode) reject(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 3 {
		return shim.Error("票据背书拒绝签收失败, 参数个数错误")
	}
	bill, bl := GetBill(stub, args[0])
	if !bl {
		return shim.Error("票据背书拒绝签收失败, 根据票据号码查询信息时发生错误")
	}

	// 以待背书人证件号码及票据号码创建复合键, 以便于当前用户无法再从待背书票据列表中查询到该票据信息
	waitEndorseCmIDBillInfoIDIndexKey, err := stub.CreateCompositeKey(IndexName, []string{args[1], bill.BillInfoID})
	if err != nil {
		return shim.Error("票据背书拒绝签收失败, 构建复合键时发生错误")
	}
	err = stub.DelState(waitEndorseCmIDBillInfoIDIndexKey)
	if err != nil {
		return shim.Error("票据背书拒绝签收失败, 删除复合键时发生错误")
	}

	// 修改票据信息: 票据状态, 待背书人信息, 拒绝背书人信息
	bill.State = BillInfo_State_EndorseReject
	bill.RejectEndorseAcct = args[2]
	bill.RejectEndorseCmID = args[1]
	bill.WaitEndorseAcct = ""
	bill.WaitEndorseCmID = ""

	// 保存票据状态
	_, bl = PutBill(stub, bill)
	if !bl {
		return shim.Error("票据背书拒绝签收失败, 保存票据状态时发生错误")
	}

	/*msg, _ := GetMsgByte(0, "票据背书拒签成功")
	return shim.Success(msg)*/
	return shim.Success([]byte("票据背书拒签成功"))
}
