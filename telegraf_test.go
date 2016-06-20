package telegraf

import (
	"math/rand"
	"testing"
	"time"
)

func TestSendMessage(h *testing.T) {
	t := NewTelegraf()
	t.SetProtocol("tcp")
	t.SetServiceAddress("118.193.80.78:18094")
	for i := 0; i < 10000; i++ {
		t.SetMeasurement("student")
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
