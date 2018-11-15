package main

import (
	"fmt"
	"github.com/kongyixueyuan.com/bill/blockchain"
	"github.com/kongyixueyuan.com/bill/service"
	"github.com/kongyixueyuan.com/bill/web"
	"github.com/kongyixueyuan.com/bill/web/controller"
	"os"
)

func main() {
	fsetup := blockchain.FabricSetup{
		ConfigFile: "config.yaml",
		ChannelID:  "mychannel",
		// /home/kevin/go
		ChannelConfig: os.Getenv("GOPATH") + "/src/github.com/kongyixueyuan.com/bill/fixtures/artifacts/channel.tx",

		OrgAdmin: "Admin",
		OrgName:  "Org1",

		// 链码相关
		ChaincodeID:     "bill",
		ChaincodeGoPath: os.Getenv("GOPATH"),
		ChaincodePath:   "github.com/kongyixueyuan.com/bill/chaincode/",
		UserName:        "User1",
	}

	err := fsetup.Initialize()
	if err != nil {
		fmt.Printf(err.Error())
	}

	err = fsetup.InstallAndInstantiateCC()
	if err != nil {
		fmt.Println(err.Error())
	}

	// 测试业务层调用链码
	fsetupService := new(service.FabricSetupService)
	fsetupService.Fabric = &fsetup

	/**
	//==========================  业务层测试开始  ======================================//

	// 发布票据
	bill := service.Bill{
		BillInfoID:"BOC1001",
		BillInfoAmt:"20000",
		BillInfoType:"111",
		BillInfoIsseDate:"20180101",
		BillInfoDueDate:"201801011",

		DrwrAcct:"111",
		DrwrCmID:"111",
		AccptrAcct:"111",
		AccptrCmID:"111",

		PyeeAcct:"111",
		PyeeCmID:"111",
		HoldrAcct:"jack",
		HoldrCmID:"jackID",
	}

	bill2 := service.Bill{
		BillInfoID:"BOC1002",
		BillInfoAmt:"10000",
		BillInfoType:"111",
		BillInfoIsseDate:"20180101",
		BillInfoDueDate:"201801011",

		DrwrAcct:"111",
		DrwrCmID:"111",
		AccptrAcct:"111",
		AccptrCmID:"111",

		PyeeAcct:"111",
		PyeeCmID:"111",
		HoldrAcct:"jack",
		HoldrCmID:"jackID",
	}

	msg, err := fsetupService.SaveBill(bill)
	if err != nil {
		fmt.Println(err.Error())
	}else {
		fmt.Println("票据发布成功, 交易编号为: " + msg)
	}

	msg, err = fsetupService.SaveBill(bill2)
	if err != nil {
		fmt.Println(err.Error())
	}else {
		fmt.Println("票据发布成功, 交易编号为: " + msg)
	}

	// 查询当前持票人的票据列表
	result, err := fsetupService.FindBills("jackID")
	if err != nil {
		fmt.Println(err.Error())
	}else {
		fmt.Println("根据持票人证件号码查询票据列表成功")
		var bills = []service.Bill{}
		json.Unmarshal(result, &bills)
		for _, obj := range bills{
			fmt.Println(obj)
		}
	}

	// 发起背书请求
	msg, err = fsetupService.Endorse("BOC1001", "aliceID", "alice")
	if err != nil {
		fmt.Println(err.Error())
	}else {
		fmt.Println(msg)
	}

	// 查询待背书票据列表
	result, err = fsetupService.FindWaitBills("aliceID")
	if err != nil{
		fmt.Println(err.Error())
	}else {
		fmt.Println("查询待背书票据列表成功")
		var bills = []service.Bill{}
		json.Unmarshal(result, &bills)
		for _, obj := range bills {
			fmt.Println(obj)
		}
	}

	// 签收票据
	msg, err = fsetupService.Accept("BOC1001", "aliceID", "alice")
	if err != nil {
		fmt.Println(err.Error())
	}else {
		fmt.Println(msg)
	}

	// 根据票据号码查询票据详情
	result, err = fsetupService.FindBillByNo("BOC1001")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		var bill service.Bill
		json.Unmarshal(result, &bill)
		fmt.Println(bill)
	}

	//=====================  拒签票据测试  ==============================//
	// 发起背书
	msg, err = fsetupService.Endorse("BOC1002", "aliceID", "alice")
	if err != nil {
		fmt.Println(err.Error())
	}else {
		fmt.Println(msg)
	}

	// 票据拒签
	msg, err = fsetupService.Reject("BOC1002", "aliceID", "alice")
	if err != nil {
		fmt.Println(err.Error())
	}else{
		fmt.Println(msg)
	}

	result, err = fsetupService.FindBillByNo("BOC1002")
	if err != nil {
		fmt.Println(err.Error())
	} else {
		var bill service.Bill
		json.Unmarshal(result, &bill)
		fmt.Println(bill)
	}

	//==========================  业务层测试完毕  ======================================//
	*/

	// 调用WebServer启动Web服务
	app := controller.Application{
		Setup: fsetupService,
	}
	web.WebStart(app)
}
