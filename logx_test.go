package logx

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"testing"
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

func TestMkConfigFile(t *testing.T) {
	c1 := configFile{}
	c1.OutputFlag = []string{"console", "dbgview", "file"}
	c1.OutputLevel = "debug"

	buf, _ := yaml.Marshal(c1)
	ioutil.WriteFile("log.yaml", buf, 0666)
}

func TestRdConfigFile(t *testing.T) {
	c1 := configFile{}

	buf, _ := ioutil.ReadFile("log.yaml")
	yaml.Unmarshal(buf, &c1)
	t.Log(c1)
}
