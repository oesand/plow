package prm

import (
	"context"

	"github.com/oesand/plow"
)

// ParamHandler is a handler that takes a single parameter.
func ParamHandler[T0 any](
	provider ParameterProvider[T0],
	handler func(context.Context, T0) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0)
	})
}

// ParamHandler2 is a handler that takes two parameters.
func ParamHandler2[T0 any, T1 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	handler func(context.Context, T0, T1) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1)
	})
}

// ParamHandler3 is a handler that takes three parameters.
func ParamHandler3[T0 any, T1 any, T2 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	handler func(context.Context, T0, T1, T2) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2)
	})
}

// ParamHandler4 is a handler that takes four parameters.
func ParamHandler4[T0 any, T1 any, T2 any, T3 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	provider3 ParameterProvider[T3],
	handler func(context.Context, T0, T1, T2, T3) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p3, resp := provider3.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2, p3)
	})
}

// ParamHandler5 is a handler that takes five parameters.
func ParamHandler5[T0 any, T1 any, T2 any, T3 any, T4 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	provider3 ParameterProvider[T3],
	provider4 ParameterProvider[T4],
	handler func(context.Context, T0, T1, T2, T3, T4) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p3, resp := provider3.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p4, resp := provider4.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2, p3, p4)
	})
}

// ParamHandler6 is a handler that takes six parameters.
func ParamHandler6[T0 any, T1 any, T2 any, T3 any, T4 any, T5 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	provider3 ParameterProvider[T3],
	provider4 ParameterProvider[T4],
	provider5 ParameterProvider[T5],
	handler func(context.Context, T0, T1, T2, T3, T4, T5) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p3, resp := provider3.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p4, resp := provider4.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p5, resp := provider5.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2, p3, p4, p5)
	})
}

// ParamHandler7 is a handler that takes seven parameters.
func ParamHandler7[T0 any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	provider3 ParameterProvider[T3],
	provider4 ParameterProvider[T4],
	provider5 ParameterProvider[T5],
	provider6 ParameterProvider[T6],
	handler func(context.Context, T0, T1, T2, T3, T4, T5, T6) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p3, resp := provider3.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p4, resp := provider4.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p5, resp := provider5.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p6, resp := provider6.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2, p3, p4, p5, p6)
	})
}

// ParamHandler8 is a handler that takes eight parameters.
func ParamHandler8[T0 any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	provider3 ParameterProvider[T3],
	provider4 ParameterProvider[T4],
	provider5 ParameterProvider[T5],
	provider6 ParameterProvider[T6],
	provider7 ParameterProvider[T7],
	handler func(context.Context, T0, T1, T2, T3, T4, T5, T6, T7) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p3, resp := provider3.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p4, resp := provider4.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p5, resp := provider5.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p6, resp := provider6.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p7, resp := provider7.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2, p3, p4, p5, p6, p7)
	})
}

// ParamHandler9 is a handler that takes nine parameters.
func ParamHandler9[T0 any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	provider3 ParameterProvider[T3],
	provider4 ParameterProvider[T4],
	provider5 ParameterProvider[T5],
	provider6 ParameterProvider[T6],
	provider7 ParameterProvider[T7],
	provider8 ParameterProvider[T8],
	handler func(context.Context, T0, T1, T2, T3, T4, T5, T6, T7, T8) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p3, resp := provider3.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p4, resp := provider4.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p5, resp := provider5.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p6, resp := provider6.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p7, resp := provider7.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p8, resp := provider8.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2, p3, p4, p5, p6, p7, p8)
	})
}

// ParamHandler10 is a handler that takes ten parameters.
func ParamHandler10[T0 any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	provider3 ParameterProvider[T3],
	provider4 ParameterProvider[T4],
	provider5 ParameterProvider[T5],
	provider6 ParameterProvider[T6],
	provider7 ParameterProvider[T7],
	provider8 ParameterProvider[T8],
	provider9 ParameterProvider[T9],
	handler func(context.Context, T0, T1, T2, T3, T4, T5, T6, T7, T8, T9) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p3, resp := provider3.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p4, resp := provider4.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p5, resp := provider5.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p6, resp := provider6.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p7, resp := provider7.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p8, resp := provider8.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p9, resp := provider9.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2, p3, p4, p5, p6, p7, p8, p9)
	})
}

// ParamHandler11 is a handler that takes eleven parameters.
func ParamHandler11[T0 any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	provider3 ParameterProvider[T3],
	provider4 ParameterProvider[T4],
	provider5 ParameterProvider[T5],
	provider6 ParameterProvider[T6],
	provider7 ParameterProvider[T7],
	provider8 ParameterProvider[T8],
	provider9 ParameterProvider[T9],
	provider10 ParameterProvider[T10],
	handler func(context.Context, T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p3, resp := provider3.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p4, resp := provider4.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p5, resp := provider5.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p6, resp := provider6.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p7, resp := provider7.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p8, resp := provider8.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p9, resp := provider9.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p10, resp := provider10.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2, p3, p4, p5, p6, p7, p8, p9, p10)
	})
}

// ParamHandler12 is a handler that takes twelve parameters.
func ParamHandler12[T0 any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	provider3 ParameterProvider[T3],
	provider4 ParameterProvider[T4],
	provider5 ParameterProvider[T5],
	provider6 ParameterProvider[T6],
	provider7 ParameterProvider[T7],
	provider8 ParameterProvider[T8],
	provider9 ParameterProvider[T9],
	provider10 ParameterProvider[T10],
	provider11 ParameterProvider[T11],
	handler func(context.Context, T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p3, resp := provider3.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p4, resp := provider4.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p5, resp := provider5.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p6, resp := provider6.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p7, resp := provider7.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p8, resp := provider8.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p9, resp := provider9.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p10, resp := provider10.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p11, resp := provider11.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2, p3, p4, p5, p6, p7, p8, p9, p10, p11)
	})
}

// ParamHandler13 is a handler that takes thirteen parameters.
func ParamHandler13[T0 any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	provider3 ParameterProvider[T3],
	provider4 ParameterProvider[T4],
	provider5 ParameterProvider[T5],
	provider6 ParameterProvider[T6],
	provider7 ParameterProvider[T7],
	provider8 ParameterProvider[T8],
	provider9 ParameterProvider[T9],
	provider10 ParameterProvider[T10],
	provider11 ParameterProvider[T11],
	provider12 ParameterProvider[T12],
	handler func(context.Context, T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p3, resp := provider3.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p4, resp := provider4.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p5, resp := provider5.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p6, resp := provider6.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p7, resp := provider7.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p8, resp := provider8.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p9, resp := provider9.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p10, resp := provider10.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p11, resp := provider11.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p12, resp := provider12.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2, p3, p4, p5, p6, p7, p8, p9, p10, p11, p12)
	})
}

// ParamHandler14 is a handler that takes fourteen parameters.
func ParamHandler14[T0 any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	provider3 ParameterProvider[T3],
	provider4 ParameterProvider[T4],
	provider5 ParameterProvider[T5],
	provider6 ParameterProvider[T6],
	provider7 ParameterProvider[T7],
	provider8 ParameterProvider[T8],
	provider9 ParameterProvider[T9],
	provider10 ParameterProvider[T10],
	provider11 ParameterProvider[T11],
	provider12 ParameterProvider[T12],
	provider13 ParameterProvider[T13],
	handler func(context.Context, T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p3, resp := provider3.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p4, resp := provider4.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p5, resp := provider5.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p6, resp := provider6.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p7, resp := provider7.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p8, resp := provider8.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p9, resp := provider9.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p10, resp := provider10.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p11, resp := provider11.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p12, resp := provider12.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p13, resp := provider13.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2, p3, p4, p5, p6, p7, p8, p9, p10, p11, p12, p13)
	})
}

// ParamHandler15 is a handler that takes fifteen parameters.
func ParamHandler15[T0 any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	provider3 ParameterProvider[T3],
	provider4 ParameterProvider[T4],
	provider5 ParameterProvider[T5],
	provider6 ParameterProvider[T6],
	provider7 ParameterProvider[T7],
	provider8 ParameterProvider[T8],
	provider9 ParameterProvider[T9],
	provider10 ParameterProvider[T10],
	provider11 ParameterProvider[T11],
	provider12 ParameterProvider[T12],
	provider13 ParameterProvider[T13],
	provider14 ParameterProvider[T14],
	handler func(context.Context, T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p3, resp := provider3.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p4, resp := provider4.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p5, resp := provider5.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p6, resp := provider6.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p7, resp := provider7.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p8, resp := provider8.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p9, resp := provider9.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p10, resp := provider10.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p11, resp := provider11.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p12, resp := provider12.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p13, resp := provider13.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p14, resp := provider14.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2, p3, p4, p5, p6, p7, p8, p9, p10, p11, p12, p13, p14)
	})
}

// ParamHandler16 is a handler that takes sixteen parameters.
func ParamHandler16[T0 any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	provider3 ParameterProvider[T3],
	provider4 ParameterProvider[T4],
	provider5 ParameterProvider[T5],
	provider6 ParameterProvider[T6],
	provider7 ParameterProvider[T7],
	provider8 ParameterProvider[T8],
	provider9 ParameterProvider[T9],
	provider10 ParameterProvider[T10],
	provider11 ParameterProvider[T11],
	provider12 ParameterProvider[T12],
	provider13 ParameterProvider[T13],
	provider14 ParameterProvider[T14],
	provider15 ParameterProvider[T15],
	handler func(context.Context, T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p3, resp := provider3.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p4, resp := provider4.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p5, resp := provider5.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p6, resp := provider6.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p7, resp := provider7.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p8, resp := provider8.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p9, resp := provider9.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p10, resp := provider10.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p11, resp := provider11.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p12, resp := provider12.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p13, resp := provider13.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p14, resp := provider14.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p15, resp := provider15.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2, p3, p4, p5, p6, p7, p8, p9, p10, p11, p12, p13, p14, p15)
	})
}

// ParamHandler17 is a handler that takes seventeen parameters.
func ParamHandler17[T0 any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	provider3 ParameterProvider[T3],
	provider4 ParameterProvider[T4],
	provider5 ParameterProvider[T5],
	provider6 ParameterProvider[T6],
	provider7 ParameterProvider[T7],
	provider8 ParameterProvider[T8],
	provider9 ParameterProvider[T9],
	provider10 ParameterProvider[T10],
	provider11 ParameterProvider[T11],
	provider12 ParameterProvider[T12],
	provider13 ParameterProvider[T13],
	provider14 ParameterProvider[T14],
	provider15 ParameterProvider[T15],
	provider16 ParameterProvider[T16],
	handler func(context.Context, T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p3, resp := provider3.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p4, resp := provider4.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p5, resp := provider5.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p6, resp := provider6.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p7, resp := provider7.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p8, resp := provider8.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p9, resp := provider9.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p10, resp := provider10.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p11, resp := provider11.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p12, resp := provider12.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p13, resp := provider13.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p14, resp := provider14.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p15, resp := provider15.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p16, resp := provider16.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2, p3, p4, p5, p6, p7, p8, p9, p10, p11, p12, p13, p14, p15, p16)
	})
}

// ParamHandler18 is a handler that takes eighteen parameters.
func ParamHandler18[T0 any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	provider3 ParameterProvider[T3],
	provider4 ParameterProvider[T4],
	provider5 ParameterProvider[T5],
	provider6 ParameterProvider[T6],
	provider7 ParameterProvider[T7],
	provider8 ParameterProvider[T8],
	provider9 ParameterProvider[T9],
	provider10 ParameterProvider[T10],
	provider11 ParameterProvider[T11],
	provider12 ParameterProvider[T12],
	provider13 ParameterProvider[T13],
	provider14 ParameterProvider[T14],
	provider15 ParameterProvider[T15],
	provider16 ParameterProvider[T16],
	provider17 ParameterProvider[T17],
	handler func(context.Context, T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16, T17) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p3, resp := provider3.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p4, resp := provider4.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p5, resp := provider5.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p6, resp := provider6.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p7, resp := provider7.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p8, resp := provider8.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p9, resp := provider9.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p10, resp := provider10.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p11, resp := provider11.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p12, resp := provider12.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p13, resp := provider13.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p14, resp := provider14.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p15, resp := provider15.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p16, resp := provider16.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p17, resp := provider17.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2, p3, p4, p5, p6, p7, p8, p9, p10, p11, p12, p13, p14, p15, p16, p17)
	})
}

// ParamHandler19 is a handler that takes nineteen parameters.
func ParamHandler19[T0 any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	provider3 ParameterProvider[T3],
	provider4 ParameterProvider[T4],
	provider5 ParameterProvider[T5],
	provider6 ParameterProvider[T6],
	provider7 ParameterProvider[T7],
	provider8 ParameterProvider[T8],
	provider9 ParameterProvider[T9],
	provider10 ParameterProvider[T10],
	provider11 ParameterProvider[T11],
	provider12 ParameterProvider[T12],
	provider13 ParameterProvider[T13],
	provider14 ParameterProvider[T14],
	provider15 ParameterProvider[T15],
	provider16 ParameterProvider[T16],
	provider17 ParameterProvider[T17],
	provider18 ParameterProvider[T18],
	handler func(context.Context, T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16, T17, T18) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p3, resp := provider3.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p4, resp := provider4.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p5, resp := provider5.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p6, resp := provider6.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p7, resp := provider7.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p8, resp := provider8.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p9, resp := provider9.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p10, resp := provider10.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p11, resp := provider11.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p12, resp := provider12.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p13, resp := provider13.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p14, resp := provider14.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p15, resp := provider15.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p16, resp := provider16.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p17, resp := provider17.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p18, resp := provider18.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2, p3, p4, p5, p6, p7, p8, p9, p10, p11, p12, p13, p14, p15, p16, p17, p18)
	})
}

// ParamHandler20 is a handler that takes twenty parameters.
func ParamHandler20[T0 any, T1 any, T2 any, T3 any, T4 any, T5 any, T6 any, T7 any, T8 any, T9 any, T10 any, T11 any, T12 any, T13 any, T14 any, T15 any, T16 any, T17 any, T18 any, T19 any](
	provider ParameterProvider[T0],
	provider1 ParameterProvider[T1],
	provider2 ParameterProvider[T2],
	provider3 ParameterProvider[T3],
	provider4 ParameterProvider[T4],
	provider5 ParameterProvider[T5],
	provider6 ParameterProvider[T6],
	provider7 ParameterProvider[T7],
	provider8 ParameterProvider[T8],
	provider9 ParameterProvider[T9],
	provider10 ParameterProvider[T10],
	provider11 ParameterProvider[T11],
	provider12 ParameterProvider[T12],
	provider13 ParameterProvider[T13],
	provider14 ParameterProvider[T14],
	provider15 ParameterProvider[T15],
	provider16 ParameterProvider[T16],
	provider17 ParameterProvider[T17],
	provider18 ParameterProvider[T18],
	provider19 ParameterProvider[T19],
	handler func(context.Context, T0, T1, T2, T3, T4, T5, T6, T7, T8, T9, T10, T11, T12, T13, T14, T15, T16, T17, T18, T19) plow.Response,
) plow.Handler {
	return plow.HandlerFunc(func(ctx context.Context, request plow.Request) plow.Response {
		p0, resp := provider.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p1, resp := provider1.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p2, resp := provider2.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p3, resp := provider3.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p4, resp := provider4.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p5, resp := provider5.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p6, resp := provider6.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p7, resp := provider7.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p8, resp := provider8.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p9, resp := provider9.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p10, resp := provider10.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p11, resp := provider11.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p12, resp := provider12.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p13, resp := provider13.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p14, resp := provider14.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p15, resp := provider15.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p16, resp := provider16.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p17, resp := provider17.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p18, resp := provider18.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		p19, resp := provider19.GetParamValue(ctx, request)
		if resp != nil {
			return resp
		}
		return handler(ctx, p0, p1, p2, p3, p4, p5, p6, p7, p8, p9, p10, p11, p12, p13, p14, p15, p16, p17, p18, p19)
	})
}
