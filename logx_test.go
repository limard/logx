package logx

import "testing"

func TestName(t *testing.T) {
	Info("eiofdghjkkmk")
	Info("eiofdghjkkmk\r")
	Info("eiofdghjkkmk\n")
	Info("eiofdghjkkmk\r\n")
}
