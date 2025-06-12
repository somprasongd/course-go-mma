package mediator

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

// ใช้แทนกรณีไม่ต้องการ response ใด ๆ
type NoResponse struct{}

// Interface สำหรับ handler ที่รับ request และ return response
type RequestHandler[TRequest any, TResponse any] interface {
	Handle(ctx context.Context, request TRequest) (TResponse, error)
}

// registry สำหรับเก็บ handler ตาม type ของ request
var handlers = map[reflect.Type]func(ctx context.Context, req interface{}) (interface{}, error){}

// Register: ผูก handler กับ type ของ request ที่รองรับ
func Register[TRequest any, TResponse any](handler RequestHandler[TRequest, TResponse]) {
	var req TRequest // สร้าง zero value เพื่อใช้หา type
	reqType := reflect.TypeOf(req)

	// wrap handler ให้รองรับ interface{}
	handlers[reqType] = func(ctx context.Context, request interface{}) (interface{}, error) {
		typedReq, ok := request.(TRequest)
		if !ok {
			return nil, errors.New("invalid request type")
		}
		return handler.Handle(ctx, typedReq)
	}
}

// Send: dispatch request ไปยัง handler ที่ match กับ type ของ request
func Send[TRequest any, TResponse any](ctx context.Context, req TRequest) (TResponse, error) {
	reqType := reflect.TypeOf(req)
	handler, ok := handlers[reqType]
	if !ok {
		var empty TResponse
		return empty, fmt.Errorf("no handler for request %T", req)
	}

	result, err := handler(ctx, req)
	if err != nil {
		var empty TResponse
		return empty, err
	}

	// ตรวจสอบ type ของ response ก่อน return
	typedRes, ok := result.(TResponse)
	if !ok {
		var empty TResponse
		return empty, errors.New("invalid response type")
	}

	return typedRes, nil
}
