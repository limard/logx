package logx

import (
	"io/ioutil"
	"testing"
	"encoding/json"
)

func TestName(t *testing.T) {
	SetOutputFlag(OutputFlag_File)

	Info("eiofdghjkkmk")

	//for i:=0; i<100000; i++ {
	//	Info("eiofdghjkkmk")
	//	Info("eiofdghjkkmk\r")
	//	Info("eiofdghjkkmk\n")
	//	Info("eiofdghjkkmk\r\n")
	//}
}

func TestConfigFile(t *testing.T) {
	c1 := configFile{}
	c1.OutputFlag = []string{"console", "dbgview", "file"}
	c1.OutputLevel = "debug"

	buf, _ := json.Marshal(c1)
	ioutil.WriteFile("log.json", buf, 0666)

	buf, _ = ioutil.ReadFile("log.json")
	json.Unmarshal(buf, &c1)
	t.Log(c1)
}
