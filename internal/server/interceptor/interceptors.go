package interceptor

import (
	"connectrpc.com/connect"
	"connectrpc.com/validate"
)

func NewInterceptors() []connect.Interceptor {
	return []connect.Interceptor{
		validate.NewInterceptor(),
	}
}

func Option() connect.Option {
	return connect.WithInterceptors(NewInterceptors()...)
}
