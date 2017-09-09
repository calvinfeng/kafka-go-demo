package main

import "encoding/json"

type accessLogEntry struct {
	Method       string  `json:"method"`
	Host         string  `json:"host"`
	Path         string  `json:"path"`
	IP           string  `json:"ip"`
	ResponseTime float64 `json:"response_time"`

	encoded []byte
	err     error
}

func (ale *accessLogEntry) ensureEncoded() {
	if ale.encoded == nil && ale.err == nil {
		ale.encoded, ale.err = json.Marshal(ale)
	}
}

func (ale *accessLogEntry) Length() int {
	ale.ensureEncoded()
	return len(ale.encoded)
}

func (ale *accessLogEntry) Encode() ([]byte, error) {
	ale.ensureEncoded()
	return ale.encoded, ale.err
}