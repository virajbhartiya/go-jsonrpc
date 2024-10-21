package jsonrpc

import (
	"encoding/json"
	"errors"
	"reflect"
)

const eTempWSError = -1111111

type RPCConnectionError struct {
	err error
}

func (e *RPCConnectionError) Error() string {
	if e.err != nil {
		return e.err.Error()
	}
	return "RPCConnectionError"
}

func (e *RPCConnectionError) Unwrap() error {
	if e.err != nil {
		return e.err
	}
	return errors.New("RPCConnectionError")
}

type ErrorCode int

type ErrorConstructor func(message string, data json.RawMessage, meta json.RawMessage) error

type Errors struct {
	byType map[reflect.Type]ErrorCode
	byCode map[ErrorCode]ErrorConstructor
}

const FirstUserCode = 2

func NewErrors() Errors {
	return Errors{
		byType: map[reflect.Type]ErrorCode{},
		byCode: map[ErrorCode]ErrorConstructor{
			-1111111: func(message string, data json.RawMessage, meta json.RawMessage) error {
				return &RPCConnectionError{}
			},
		},
	}
}

func (e *Errors) Register(c ErrorCode, typ interface{}, constructor ErrorConstructor) {
	rt := reflect.TypeOf(typ).Elem()
	if !rt.Implements(errorType) {
		panic("can't register non-error types")
	}

	e.byType[rt] = c
	if constructor != nil {
		e.byCode[c] = constructor
	} else {
		e.byCode[c] = func(message string, data json.RawMessage, meta json.RawMessage) error {
			var v reflect.Value
			if rt.Kind() == reflect.Ptr {
				v = reflect.New(rt.Elem())
			} else {
				v = reflect.New(rt)
			}
			if len(meta) > 0 && v.Type().Implements(marshalableRT) {
				_ = v.Interface().(marshalable).UnmarshalJSON(meta)
			}
			if rt.Kind() != reflect.Ptr {
				v = v.Elem()
			}
			return v.Interface().(error)
		}
	}
}

type marshalable interface {
	json.Marshaler
	json.Unmarshaler
}

// ErrorWithData contains extra data to explain the error
type ErrorWithData interface {
	Error() string  // returns the message
	ErrorData() any // returns the error data
}
