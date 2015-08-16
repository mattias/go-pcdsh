package main

type Configuration struct {
	BaseUrl    string
	Port       string
	Datasource string
}

type RespStruct struct {
	Result   string
	Response map[string]interface{}
}
