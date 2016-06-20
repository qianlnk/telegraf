package telegraf

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

type Telegraf struct {
	protocol       string
	serviceAddress string

	sync.Mutex

	measurement string
	tags        map[string]string
	values      map[string]interface{}
	timestamp   int64

	conn  net.Conn
	msgCh chan []byte
}

func NewTelegraf() *Telegraf {
	return &Telegraf{
		tags:   make(map[string]string),
		values: make(map[string]interface{}),
		msgCh:  make(chan []byte),
	}
}

func (t *Telegraf) read() {
	for {
		buf := make([]byte, 1024)
		n, err := t.conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			break
		}
		fmt.Println(string(buf[1:n]))
	}
}

func (t *Telegraf) keeepConnect() {
	for {
		var err error
		t.conn, err = net.Dial(t.protocol, t.serviceAddress)
		if err != nil {
			fmt.Println(err)
			time.Sleep(5 * time.Second)
			continue
		}
		go t.read()
		for {
			errCount := 0
			select {
			case msg := <-t.msgCh:
				fmt.Println(string(msg))
				_, err := t.conn.Write(msg)
				if err != nil {
					fmt.Println(err)
					t.msgCh <- msg //if err write back the msg
					errCount++
				}
			}
			if errCount != 0 {
				break
			}
		}
	}
}

func (t *Telegraf) SetProtocol(protocol string) {
	t.protocol = protocol
}

func (t *Telegraf) SetServiceAddress(address string) {
	t.serviceAddress = address
	go t.keeepConnect()
	//time.Sleep(1 * time.Second)
}

func (t *Telegraf) SetMeasurement(measurement string) {
	t.measurement = measurement
}

func (t *Telegraf) AddTag(tagName, tagValue string) error {
	t.Lock()
	defer t.Unlock()
	if _, ok := t.tags[tagName]; ok {
		return errors.New(fmt.Sprintf("tag: %s is exist.", tagName))
	}
	t.tags[tagName] = tagValue
	return nil
}

func (t *Telegraf) AddValue(field string, value interface{}) error {
	t.Lock()
	defer t.Unlock()
	if _, ok := t.values[field]; ok {
		return errors.New(fmt.Sprintf("field: %s is exist.", field))
	}
	t.values[field] = value
	return nil
}

func (t *Telegraf) SetTimestamp(timing time.Time) {
	t.timestamp = timing.UnixNano()
	//t.timestamp = timing.Unix()
}

func (t *Telegraf) getMessage() (string, error) {
	var message string
	message = t.measurement
	for tag, value := range t.tags {
		value = strings.Replace(value, " ", "\\ ", -1)
		value = strings.Replace(value, ",", "\\,", -1)
		message += fmt.Sprintf(",%s=%s", tag, value)
	}

	fieldNum := 0
	for field, value := range t.values {
		if fieldNum == 0 {
			message += " "
		} else {
			message += ","
		}
		switch value.(type) {
		case int:
			message += fmt.Sprintf("%s=%di", field, value.(int))
			break
		case int64:
			message += fmt.Sprintf("%s=%di", field, value.(int64))
			break
		case float64:
			message += fmt.Sprintf("%s=%f", field, value.(float64))
			break
		case bool:
			if value.(bool) == true {
				message += fmt.Sprintf("%s=true", field)
			} else {
				message += fmt.Sprintf("%s=false", field)
			}
			break
		case string:
			message += fmt.Sprintf("%s=\"%s\"", field, value.(string))
			break
		default:
			return "", errors.New("no sush type")
		}
		fieldNum++
	}

	if t.timestamp != 0 {
		message += fmt.Sprintf(" %d", t.timestamp)
	}
	message += "\n"
	return message, nil
}

func (t *Telegraf) clean() {
	t.measurement = ""
	for k := range t.tags {
		delete(t.tags, k)
	}
	//t.tags = nil
	for k := range t.values {
		delete(t.values, k)
	}
	//t.values = nil
}

func (t *Telegraf) Send() error {
	t.Lock()
	defer t.Unlock()

	message, err := t.getMessage()
	if err != nil {
		return err
	}

	t.msgCh <- []byte(message)
	t.clean()

	return nil
}
