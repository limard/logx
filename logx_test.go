package logx

import "testing"

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
