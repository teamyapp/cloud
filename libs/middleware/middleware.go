package middleware

type Middleware[Executor any] func(executor Executor) Executor

func WithMiddlewares[Executor any](executor Executor, middlewares []Middleware[Executor]) Executor {
	if len(middlewares) == 0 {
		return executor
	}

	return middlewares[0](WithMiddlewares(executor, middlewares[1:]))
}
