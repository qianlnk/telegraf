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
}

func NewTelegraf() *Telegraf {
	return &Telegraf{
		tags:   make(map[string]string),
		values: make(map[string]interface{}),
	}
}

func (t *Telegraf) SetProtocol(protocol string) {
	t.protocol = protocol
}

func (t *Telegraf) SetServiceAddress(address string) {
	t.serviceAddress = address
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
	//t.timestamp = timing.UnixNano()
	t.timestamp = timing.Unix()
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

	conn, err := net.Dial(t.protocol, t.serviceAddress)
	if err != nil {
		return err
	}

	message, err := t.getMessage()
	if err != nil {
		return err
	}
	fmt.Println(message)
	fmt.Fprintf(conn, message)
	time.Sleep(time.Millisecond * 15)
	conn.Close()
	t.clean()
	return nil
}
