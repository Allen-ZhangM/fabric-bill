package controller

import (
	"encoding/json"
	"fmt"
	"github.com/kongyixueyuan.com/bill/service"
	"net/http"
)

var cuser User

func (app *Application) LoginView(w http.ResponseWriter, r *http.Request) {

	ShowView(w, r, "login.html", nil)
}

// 用户登录
func (app *Application) Login(w http.ResponseWriter, r *http.Request) {
	loginName := r.FormValue("loginName")
	password := r.FormValue("password")

	var flag bool
	for _, user := range users {
		if user.LoginName == loginName && user.Password == password {
			cuser = user
			flag = true
			break
		}
	}

	data := &struct {
		CurrentUser User
		Flag        bool
	}{
		Flag: false,
	}

	if flag {
		// 登录成功
		r.Form.Set("holdrCmID", cuser.CmID)
		app.FindBills(w, r)
	} else {
		// 登录失败
		data.Flag = true
		data.CurrentUser.LoginName = loginName
		ShowView(w, r, "login.html", data)
	}
}

// 显示发布票据页面
func (app *Application) IssueShow(w http.ResponseWriter, r *http.Request) {
	data := &struct {
		CurrentUser User
		Msg         string
		Flag        bool
	}{
		CurrentUser: cuser,
		Msg:         "",
		Flag:        false,
	}
	ShowView(w, r, "issue.html", data)
}

// 发布票据
func (app *Application) Issue(w http.ResponseWriter, r *http.Request) {
	bill := service.Bill{
		BillInfoID:       r.FormValue("billInfoID"),
		BillInfoAmt:      r.FormValue("billInfoAmt"),
		BillInfoType:     r.FormValue("billInfoType"),
		BillInfoIsseDate: r.FormValue("billInfoIsseDate"),
		BillInfoDueDate:  r.FormValue("billInfoDueDate"),
		DrwrAcct:         r.FormValue("drwrAcct"),
		DrwrCmID:         r.FormValue("drwrCmID"),
		AccptrAcct:       r.FormValue("accptrAcct"),
		AccptrCmID:       r.FormValue("accptrCmID"),
		PyeeAcct:         r.FormValue("pyeeAcct"),
		PyeeCmID:         r.FormValue("pyeeCmID"),
		HoldrAcct:        r.FormValue("holdrAcct"),
		HoldrCmID:        r.FormValue("holdrCmID"),
	}

	transactionID, err := app.Setup.SaveBill(bill)

	data := &struct {
		CurrentUser User
		Msg         string
		Flag        bool
	}{
		CurrentUser: cuser,
		Flag:        true,
		Msg:         "",
	}

	if err != nil {
		data.Msg = err.Error()
	} else {
		data.Msg = "票据发布成功:" + transactionID
	}

	ShowView(w, r, "issue.html", data)

}

// 查询持票人的票据列表
func (app *Application) FindBills(w http.ResponseWriter, r *http.Request) {
	//r.FormValue("holdrCmID")
	result, err := app.Setup.FindBills(cuser.CmID)
	if err != nil {
		fmt.Println(err.Error())
	}

	var bills = []service.Bill{}
	json.Unmarshal(result, &bills)

	data := &struct {
		Bills       []service.Bill
		CurrentUser User
	}{
		Bills:       bills,
		CurrentUser: cuser,
	}

	ShowView(w, r, "bills.html", data)
}

// 根据票据号码查询票据详情
func (app *Application) BillInfoByNo(w http.ResponseWriter, r *http.Request) {
	billNo := r.FormValue("billNo")
	result, _ := app.Setup.FindBillByNo(billNo)
	var bill = service.Bill{}
	json.Unmarshal(result, &bill)

	data := &struct {
		Bill        service.Bill
		CurrentUser User
		Msg         string
		Flag        bool
	}{
		Bill:        bill,
		CurrentUser: cuser,
		Msg:         "",
		Flag:        false,
	}

	ShowView(w, r, "billInfo.html", data)
}
