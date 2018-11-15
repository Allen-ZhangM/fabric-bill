package main

import (
	"encoding/json"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"github.com/hyperledger/fabric/protos/peer"
)

// 保存票据
// args: bill
func PutBill(stub shim.ChaincodeStubInterface, bill Bill) ([]byte, bool) {
	// 将票据对象进行序列化
	b, err := json.Marshal(bill)
	if err != nil {
		return nil, false
	}

	// 保存票据状态
	err = stub.PutState(Bill_Prefix+bill.BillInfoID, b)
	if err != nil {
		return nil, false
	}

	return b, true
}

// 根据指定的票据号码查询票据状态
// args: billNo
func GetBill(stub shim.ChaincodeStubInterface, billNo string) (Bill, bool) {
	var bill Bill
	// 根据票据号码查询票据状态
	b, err := stub.GetState(Bill_Prefix + billNo)
	if err != nil {
		return bill, false
	}

	if b == nil {
		return bill, false
	}

	// 对查询到的票据状态进行反序列化
	err = json.Unmarshal(b, &bill)
	if err != nil {
		return bill, false
	}

	// 返回结果
	return bill, true
}

// 发布票据
// args:billObject
func (t *BillChaincode) issue(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	//if len(args) != 13 {
	if len(args) != 1 {
		msg, _ := GetMsgString(1, "接收到的参数必须为票据对象")
		return shim.Error(msg)
	}

	/*bill := Bill{
		BillInfoID:args[0],
		BillInfoAmt:args[1],
		BillInfoType:args[2],
		BillInfoIsseDate:args[3],
		BillInfoDueDate:args[4],
		DrwrAcct:args[5],
		DrwrCmID:args[6],
		AccptrAcct:args[7],
		AccptrCmID:args[8],
		PyeeAcct:args[9],
		PyeeCmID:args[10],
		HoldrAcct:args[11],
		HoldrCmID:args[12],
	}*/

	var bill Bill
	err := json.Unmarshal([]byte(args[0]), &bill)
	if err != nil {
		msg, _ := GetMsgString(1, "反序列化票据对象时发生错误")
		return shim.Error(msg)
	}

	// 票据号码必须唯一
	// 查重(根据票据号码进行查询)
	_, exist := GetBill(stub, bill.BillInfoID)
	if exist {
		msg, _ := GetMsgString(1, "要发布新票据已存在(票据号码重复)")
		return shim.Error(msg)
	}

	// 更改当前新发布的票据状态
	bill.State = BillInfo_State_NewPublish

	// 保存票据
	_, bl := PutBill(stub, bill)
	if !bl {
		msg, _ := GetMsgString(1, "保存票据信息时发生错误")
		return shim.Error(msg)
	}

	// 创建复合键(根据当前持票人证件号码+票据号码), 以便于后期批量查询

	/**	A
	// 当前持票人
	//qw = holdrCmID~BillNo, AAA+BOC100001
	qw2 = holdrCmID~BillNo, AAA+BOC100002
	qw3 = holdrCmID~BillNo, AAA+BOC100003
	qw4 = holdrCmID~BillNo, BBB+BOC100006

	// 待背书人
	wait = holdrCmID~BillNo, BBB+BOC100001
	wait2 = holdrCmID~BillNo, BBB+BOC100002
	wait3 = holdrCmID~BillNo, AAA+BOC100006
	*/

	holderCmIDBillInfoNoIndexKey, err := stub.CreateCompositeKey(IndexName, []string{bill.HoldrCmID, bill.BillInfoID})
	if err != nil {
		msg, _ := GetMsgString(1, "保存票据状态后, 创建对应的复合键时发生错误")
		return shim.Error(msg)
	}

	// 如果保存复合Key时指定的value为nil, 会导致后期查询不到相应的信息
	err = stub.PutState(holderCmIDBillInfoNoIndexKey, []byte{0x00})
	if err != nil {
		msg, _ := GetMsgString(1, "保存复合键时发生错误")
		return shim.Error(msg)
	}

	// 返回成功
	/*msg, _ := GetMsgByte(0, "票据发布成功")
	return shim.Success(msg)*/
	return shim.Success([]byte("票据发布成功"))
}

// 批量查询持票人拥有的票据列表
// args: HoldrCmID
func (t *BillChaincode) queryBills(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		msg, _ := GetMsgString(1, "查询票据列表失败, 必须且只能指定持票人的证件号码")
		return shim.Error(msg)
	}

	// 根据指定的持票人证件号码查询相应的票据列表
	// 根据当前持票人证件号码从已创建的复合键中查询所有的票据号码
	billIterator, err := stub.GetStateByPartialCompositeKey(IndexName, []string{args[0]})
	if err != nil {
		msg, _ := GetMsgString(1, "查询票据列表失败, 根据持票人证件号码查询所持有的票据号码时发生错误")
		return shim.Error(msg)
	}
	defer billIterator.Close()

	var bills []Bill
	// 迭代处理
	for billIterator.HasNext() {
		// k: compositeKey; v: []byte{0x00}
		kv, err := billIterator.Next()

		// v = bill.HoldrCmID|bill.BillInfoID
		if err != nil {
			return shim.Error("查询票据失败, 获取复合键时发生错误")
		}
		// [bi.ll.HoldrCmID, bill.BillInfoID]
		_, composites, err := stub.SplitCompositeKey(kv.Key)
		if err != nil {
			return shim.Error("查询票据失败, 分割复合键时发生错误")
		}
		// 从分割后的复合键中获取对应的票据号码, 然后查询相应的票据信息
		bill, bl := GetBill(stub, composites[1])
		if !bl {
			return shim.Error("根据获取到的票据号码查询相应的票据状态时发生错误")
		}

		/**

		qw = holdrCmID~BillNo, AAA+BOC100001
		qw2 = holdrCmID~BillNo, AAA+BOC100002
		qw3 = holdrCmID~BillNo, AAA+BOC100003

		wait3 = holdrCmID~BillNo, AAA+BOC100006
		*/
		// 判断当前票据必须为持有人
		// 不需要考虑当前的票据的状态
		// 判断待背书人证件号码
		/*if bill.HoldrCmID == args[0]{
			bills = append(bills, bill)
		}*/
		if bill.WaitEndorseCmID == args[0] {
			continue
		}

		bills = append(bills, bill)

	}

	// 序列化结果
	b, err := json.Marshal(bills)
	if err != nil {
		return shim.Error("查询票据失败, 序列化票据状态时发生错误")
	}
	return shim.Success(b)
}

// 根据票据号码查询票据详情
// args: billNo
func (t *BillChaincode) queryBillByNo(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("必须且只能指定要查询的票据号码")
	}

	// 查询
	bill, bl := GetBill(stub, args[0])
	if !bl {
		return shim.Error("根据指定的票据号码查询票据失败")
	}

	// 获取历史变更数据
	iterator, err := stub.GetHistoryForKey(Bill_Prefix + bill.BillInfoID)
	if err != nil {
		return shim.Error("根据指定的票据号码查询对应的历史变更数据失败")
	}
	defer iterator.Close()

	// 迭代处理
	var historys []HistoryItem
	var hisBill Bill
	for iterator.HasNext() {
		hisData, err := iterator.Next()
		if err != nil {
			return shim.Error("获取票据的的历史变更数据失败")
		}

		var historyItem HistoryItem
		historyItem.TxId = hisData.TxId
		json.Unmarshal(hisData.Value, &hisBill)

		if hisData.Value == nil {
			var empty Bill
			historyItem.Bill = empty
		} else {
			historyItem.Bill = hisBill
		}

		historys = append(historys, historyItem)

	}

	bill.Historys = historys

	// 返回
	b, err := json.Marshal(bill)
	if err != nil {
		return shim.Error("获取票据状态及背书历史失败, 序列化票据时发生错误")
	}
	return shim.Success(b)

}
