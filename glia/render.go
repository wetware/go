package glia

import (
	"context"
	"errors"
	"fmt"
	"io"
)

var ErrStatusNotSet = errors.New("status not set")

// Renderer describes types that can write a response for a given
// request.
type Renderer interface {
	// Render writes a response for r to w.
	Render(context.Context, Result) error
}

// Render is an abstraction over r.Render(w, req) for readability.
func Render(ctx context.Context, res Result, r Renderer) error {
	if err := r.Render(ctx, res); err != nil {
		return err
	}

	// Check that the renderer set a status, else the result is
	// invalid for transmission.
	if res.Status() == Status_unset {
		return ErrStatusNotSet
	}

	return nil
}

type Ok []uint64 // stack

func (stack Ok) Render(_ context.Context, res Result) error {
	defer res.SetStatus(Status_ok)

	if size := int32(len(stack)); size > 0 {
		list, err := res.NewStack(size)
		if err != nil {
			return err
		}
		for i, word := range stack {
			list.Set(i, word)
		}
	}

	return nil
}

type Failure struct {
	Stack  []uint64
	Status Status
	Err    error
}

func (f Failure) Render(res Result) error {
	defer res.SetStatus(f.Status)

	// copy stack, if there is one
	if size := int32(len(f.Stack)); size > 0 {
		list, err := res.NewStack(size)
		if err != nil {
			return err
		}
		for i, word := range f.Stack {
			list.Set(i, word)
		}
	}

	if f.Err != nil {
		// NOTE:  someone could concievably set f.Err to `errors.New("")`,
		// and this would be indistinguishable from `f.Err = nil`.
		//
		// YOLO.
		if err := res.SetInfo(f.Err.Error()); err != nil {
			return fmt.Errorf("set info: %w", err)
		}
	}

	return nil
}

func RoutingError(err error) RenderFunc {
	return func(ctx context.Context, res Result) error {
		return Failure{
			Status: Status_routingError,
			Err:    err,
		}.Render(res)
	}
}

func ProcNotFound() RenderFunc {
	return func(ctx context.Context, res Result) error {
		return Failure{
			Status: Status_procNotFound,
		}.Render(res)
	}
}

func InvalidCallStack(err error) RenderFunc {
	return func(ctx context.Context, res Result) error {
		return Failure{
			Status: Status_invalidRequest,
			Err:    err,
		}.Render(res)
	}
}

func InvalidMethod(err error) RenderFunc {
	return func(ctx context.Context, res Result) error {
		return Failure{
			Status: Status_invalidMethod,
			Err:    err,
		}.Render(res)
	}
}

func MethodNotFound() RenderFunc {
	return func(ctx context.Context, res Result) error {
		return Failure{
			Status: Status_methodNotFound,
		}.Render(res)
	}
}

func GuestError(err error) RenderFunc {
	return func(ctx context.Context, res Result) error {
		return Failure{
			Status: Status_guestError,
			Err:    err,
		}.Render(res)
	}
}

type MethodCall struct {
	P      Proc
	Method string
	Stack  []uint64
	Body   io.Reader
}

func (mc MethodCall) Render(ctx context.Context, res Result) error {
	// Acquire a lock on the process
	////
	err := mc.P.Reserve(ctx, mc.Body)
	if err != nil {
		return err
	}
	defer mc.P.Release()

	method := mc.P.Method(mc.Method)
	if method == nil {
		return Render(ctx, res, MethodNotFound())
	}

	err = method.CallWithStack(ctx, mc.Stack)
	if errors.Is(err, context.Canceled) {
		err = context.Canceled
	} else if errors.Is(err, context.DeadlineExceeded) {
		err = context.DeadlineExceeded
	}

	if err != nil {
		return Render(ctx, res, GuestError(err))
	}

	return Render(ctx, res, Ok(mc.Stack))
}

type RenderFunc func(context.Context, Result) error

func (render RenderFunc) Render(ctx context.Context, res Result) error {
	return render(ctx, res)
}
