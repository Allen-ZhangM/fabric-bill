package main

import (
	"encoding/json"
	"fmt"
)

type ChaincodeResponseMsg struct {
	Code int    // 响应消息代码(0: 代表成功,  1: 代表失败)
	dec  string // 消息具体内容/描述
}

func GetMsgByte(code int, dec string) ([]byte, error) {
	b, err := getMsg(code, dec)
	if err != nil {
		return nil, err
	}
	return b[:], nil
}

func GetMsgString(code int, dec string) (string, error) {
	b, err := getMsg(code, dec)
	if err != nil {
		return "", err
	}
	return string(b[:]), nil
}

func getMsg(code int, dec string) ([]byte, error) {
	var crm ChaincodeResponseMsg
	crm.Code = code
	crm.dec = dec

	b, err := json.Marshal(crm)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return b, nil
}
