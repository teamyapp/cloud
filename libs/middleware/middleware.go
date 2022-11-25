package middleware

type Middleware[ExecuteFunc any] func(executeFunc ExecuteFunc) ExecuteFunc

func WithMiddlewares[ExecuteFunc any](executeFunc ExecuteFunc, middlewares []Middleware[ExecuteFunc]) ExecuteFunc {
	if len(middlewares) == 0 {
		return executeFunc
	}

	return middlewares[0](WithMiddlewares(executeFunc, middlewares[1:]))
}
