package log

import "strconv"

type IFlied interface {
	Key() string
	Value() string
}

type StringField struct {
	key   string
	value string
}

func String(key, value string) StringField {
	return StringField{
		key:   key,
		value: value,
	}
}

func (s StringField) Key() string {
	return s.key
}

func (s StringField) Value() string {
	return s.value
}

type IntField struct {
	key   string
	value int
}

func Int(key string, value int) IntField {
	return IntField{
		key:   key,
		value: value,
	}
}

func (i IntField) Key() string {
	return i.key
}

func (i IntField) Value() string {
	return strconv.Itoa(i.value)
}
