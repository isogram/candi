package restserver

import (
	"crypto/tls"
	"net/http"
	"strings"

	graphqlserver "github.com/golangid/candi/codebase/app/graphql_server"
	"github.com/golangid/candi/codebase/interfaces"
	"github.com/golangid/candi/config/env"
	"github.com/golangid/candi/wrapper"
	"github.com/soheilhy/cmux"
)

type (
	option struct {
		traceMiddleware     func(http.Handler) http.Handler
		rootMiddlewares     []func(http.Handler) http.Handler
		rootHandler         http.HandlerFunc
		routerFuncs         []func(interfaces.RESTRouter)
		errorHandler        http.HandlerFunc
		httpPort            uint16
		rootPath            string
		debugMode           bool
		includeGraphQL      bool
		jaegerMaxPacketSize int
		sharedListener      cmux.CMux
		graphqlOption       graphqlserver.Option
		tlsConfig           *tls.Config
	}

	// OptionFunc type
	OptionFunc func(*option)
)

var (
	MiddlewareExcludeURLPath = map[string]struct{}{"/": {}, "/graphql": {}, "/favicon.ico": {}}
)

func getDefaultOption() option {
	return option{
		httpPort:  8000,
		rootPath:  "/",
		debugMode: true,
		rootMiddlewares: []func(http.Handler) http.Handler{
			HTTPMiddlewareCORS(
				env.BaseEnv().CORSAllowMethods, env.BaseEnv().CORSAllowHeaders,
				env.BaseEnv().CORSAllowOrigins, nil, env.BaseEnv().CORSAllowCredential,
			),
		},
		traceMiddleware: HTTPMiddlewareTracer(),
		rootHandler:     http.HandlerFunc(wrapper.HTTPHandlerDefaultRoot),
	}
}

// SetHTTPPort option func
func SetHTTPPort(port uint16) OptionFunc {
	return func(o *option) {
		o.httpPort = port
	}
}

// SetRootPath option func
func SetRootPath(rootPath string) OptionFunc {
	return func(o *option) {
		if !strings.HasPrefix(rootPath, "/") {
			rootPath = "/" + strings.Trim(rootPath, "/")
		}
		o.rootPath = rootPath
	}
}

// SetRootHTTPHandler option func
func SetRootHTTPHandler(rootHandler http.HandlerFunc) OptionFunc {
	return func(o *option) {
		o.rootHandler = rootHandler
	}
}

// SetSharedListener option func
func SetSharedListener(sharedListener cmux.CMux) OptionFunc {
	return func(o *option) {
		o.sharedListener = sharedListener
	}
}

// SetDebugMode option func
func SetDebugMode(debugMode bool) OptionFunc {
	return func(o *option) {
		o.debugMode = debugMode
	}
}

// SetIncludeGraphQL option func
func SetIncludeGraphQL(includeGraphQL bool) OptionFunc {
	return func(o *option) {
		o.includeGraphQL = includeGraphQL
	}
}

// SetJaegerMaxPacketSize option func
func SetJaegerMaxPacketSize(max int) OptionFunc {
	return func(o *option) {
		o.jaegerMaxPacketSize = max
	}
}

// SetRootMiddlewares option func
func SetRootMiddlewares(middlewares ...func(http.Handler) http.Handler) OptionFunc {
	return func(o *option) {
		o.rootMiddlewares = middlewares
	}
}

// AddRootMiddlewares option func, overide root middleware
func AddRootMiddlewares(middlewares ...func(http.Handler) http.Handler) OptionFunc {
	return func(o *option) {
		o.rootMiddlewares = append(o.rootMiddlewares, middlewares...)
	}
}

// AddGraphQLOption option func
func AddGraphQLOption(opts ...graphqlserver.OptionFunc) OptionFunc {
	return func(o *option) {
		o.graphqlOption.RootPath = "/graphql"
		for _, opt := range opts {
			opt(&o.graphqlOption)
		}
		MiddlewareExcludeURLPath[o.graphqlOption.RootPath] = struct{}{}
	}
}

// SetTLSConfig option func
func SetTLSConfig(tlsConfig *tls.Config) OptionFunc {
	return func(o *option) {
		o.tlsConfig = tlsConfig
	}
}

// SetDisableTrace option func
func SetDisableTrace() OptionFunc {
	return func(o *option) {
		o.traceMiddleware = nil
	}
}

// AddMountRouter option func
func AddMountRouter(fn func(interfaces.RESTRouter)) OptionFunc {
	return func(o *option) {
		o.routerFuncs = append(o.routerFuncs, fn)
	}
}
