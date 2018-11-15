package web

import (
	"fmt"
	"github.com/kongyixueyuan.com/bill/web/controller"
	"net/http"
)

// 启动Web服务并指定路由信息
func WebStart(app controller.Application) {

	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// 指定路由信息(匹配请求)
	http.HandleFunc("/", app.LoginView)
	http.HandleFunc("/login", app.Login)

	http.HandleFunc("/addBill", app.IssueShow) // 显示发布票据页面
	http.HandleFunc("/issue", app.Issue)       // 提交发布票据请求

	http.HandleFunc("/bills", app.FindBills)       // 查询当前持票人的票据列表
	http.HandleFunc("/billInfo", app.BillInfoByNo) // 根据票据号码查询票据详情

	fmt.Println("启动Web服务, 监听端口号为: 9000")
	err := http.ListenAndServe(":9000", nil)
	if err != nil {
		fmt.Printf("Web服务启动失败: %v", err)
	}

}

/**
<link rel="" type="" href="/static/css/style.css">
<script src="/static/js/main.js"></script>
*/
