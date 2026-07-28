package interceptor

import (
	"context"
	"runtime"

	"connectrpc.com/connect"
	"github.com/mickamy/employee-management/internal/config"
	"github.com/mickamy/employee-management/internal/lib/logger"
)

func Recovery(cfg config.App) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {

		return func(
			ctx context.Context,
			req connect.AnyRequest,
		) (resp connect.AnyResponse, err error) {
			defer func() {
				if r := recover(); r != nil {
					buf := make([]byte, 4096)
					n := runtime.Stack(buf, false)
					logger.Error(ctx, "panic recovered",
						"panic", r,
						"procedure", req.Spec().Procedure,
						"stack", string(buf[:n]),
					)
					resp = nil
					err = connect.NewError(connect.CodeInternal, nil)
				}
			}()
			return next(ctx, req)
		}
	}
}
