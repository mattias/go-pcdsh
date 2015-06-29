package main
import "time"

type Configuration struct {
	BaseUrl    string
	Datasource string
}

type RespStruct struct {
	Result   string
	Response map[string]interface{}
}

type Session struct {
	Id, StartLogId, EndLogId, LogCount int64
	StartTime, EndTime                 time.Time
}

type SessionResource struct { }
