package telegraf

import (
	"math/rand"
	"testing"
	"time"
)

func TestSendMessage(h *testing.T) {
	for i := 0; i < 10000; i++ {
		t := NewTelegraf()
		t.SetProtocol("tcp")
		t.SetServiceAddress("192.168.10.11:18094")
		t.SetMeasurement("lnkgift")
		t.AddTag("user", "qianno")
		t.AddTag("class", "jisuanji 1002")
		age := rand.Intn(100)
		t.AddValue("age", age)
		score1 := rand.Intn(100)
		score2 := rand.Float64()
		score := score2 + float64(score1)
		t.AddValue("score", score)
		//t.AddValue("name", "xiezhenjia")
		//t.AddValue("islogin", true)
		t.Send()
		time.Sleep(1 * time.Second)
	}
}
