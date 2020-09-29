package main

const moduleMainTemplate = `package {{clean $.module}}

import (
	{{isHandlerActive $.graphqlHandler}}"{{.ServiceName}}/internal/modules/{{$.module}}/delivery/graphqlhandler"
	{{isHandlerActive $.grpcHandler}}"{{.ServiceName}}/internal/modules/{{$.module}}/delivery/grpchandler"
	{{isHandlerActive $.restHandler}}"{{.ServiceName}}/internal/modules/{{$.module}}/delivery/resthandler"
	{{isHandlerActive $.isWorkerActive}}"{{.ServiceName}}/internal/modules/{{$.module}}/delivery/workerhandler"

	"pkg.agungdwiprasetyo.com/candi/codebase/factory/dependency"
	"pkg.agungdwiprasetyo.com/candi/codebase/factory/types"
	"pkg.agungdwiprasetyo.com/candi/codebase/interfaces"
)

const (
	moduleName types.Module = "{{clean (upper $.module)}}"
)

// Module model
type Module struct {
	restHandler    interfaces.EchoRestHandler
	grpcHandler    interfaces.GRPCHandler
	graphqlHandler interfaces.GraphQLHandler

	workerHandlers map[types.Worker]interfaces.WorkerHandler
}

// NewModule module constructor
func NewModule(deps dependency.Dependency) *Module {
	var mod Module
	{{isHandlerActive $.restHandler}}mod.restHandler = resthandler.NewRestHandler(deps.GetMiddleware())
	{{isHandlerActive $.grpcHandler}}mod.grpcHandler = grpchandler.NewGRPCHandler(deps.GetMiddleware())
	{{isHandlerActive $.graphqlHandler}}mod.graphqlHandler = graphqlhandler.NewGraphQLHandler(deps.GetMiddleware())

	mod.workerHandlers = map[types.Worker]interfaces.WorkerHandler{
		{{isHandlerActive $.kafkaHandler}}types.Kafka:           workerhandler.NewKafkaHandler(),
		{{isHandlerActive $.schedulerHandler}}types.Scheduler:       workerhandler.NewCronHandler(),
		{{isHandlerActive $.redissubsHandler}}types.RedisSubscriber: workerhandler.NewRedisHandler(),
		{{isHandlerActive $.taskqueueHandler}}types.TaskQueue:       workerhandler.NewTaskQueueHandler(),
	}

	return &mod
}

// RestHandler method
func (m *Module) RestHandler() interfaces.EchoRestHandler {
	return m.restHandler
}

// GRPCHandler method
func (m *Module) GRPCHandler() interfaces.GRPCHandler {
	return m.grpcHandler
}

// GraphQLHandler method
func (m *Module) GraphQLHandler() interfaces.GraphQLHandler {
	return m.graphqlHandler
}

// WorkerHandler method
func (m *Module) WorkerHandler(workerType types.Worker) interfaces.WorkerHandler {
	return m.workerHandlers[workerType]
}

// Name get module name
func (m *Module) Name() types.Module {
	return moduleName
}
`

const defaultFile = `package {{$.packageName}}`

func defaultDataSource(fileName string) []byte {
	return loadTemplate(defaultFile, map[string]string{"packageName": fileName})
}
