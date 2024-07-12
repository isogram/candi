package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/golangid/candi"
)

var (
	scopeMap = map[string]string{
		"1": InitMonorepo, "2": InitService, "3": AddModule, "4": AddHandler, "5": AddUsecase, "6": ApplyUsecase,
	}

	dependencyMap = map[string]string{
		"1": RedisDeps, "2": SqldbDeps, "3": MongodbDeps, "4": ArangodbDeps,
	}
	sqlDrivers = map[string]string{
		"1": "postgres", "2": "mysql", "3": "sqlite3",
	}
	optionYesNo = map[string]bool{"y": true, "n": false}
	licenseMap  = map[string]string{
		"1": MitLicense, "2": ApacheLicense, "3": PrivateLicense,
	}
	licenseMapTemplate = map[string]string{
		MitLicense: mitLicenseTemplate, ApacheLicense: apacheLicenseTemplate, PrivateLicense: privateLicenseTemplate,
	}

	restPluginHandler = map[string]string{
		"1": FiberRestDeps,
	}

	tpl    *template.Template
	logger *log.Logger
	reader *bufio.Reader

	specialChar        = []string{"*", "", "/", "", ":", ""}
	cleanSpecialChar   = strings.NewReplacer(append(specialChar, "-", "")...)
	modulePathReplacer = strings.NewReplacer(specialChar...)
)

type flagParameter struct {
	scopeFlag, packagePrefixFlag, protoOutputPkgFlag, outputFlag, libraryNameFlag string
	withGoModFlag                                                                 bool
	run, all                                                                      bool
	initService, addModule, addHandler, initMonorepo, version, isMonorepo         bool
	serviceName, moduleName, monorepoProjectName                                  string
	modules                                                                       []string
}

func (f *flagParameter) parseMonorepoFlag() error {
	f.packagePrefixFlag = "monorepo/services"
	f.withGoModFlag = false
	f.protoOutputPkgFlag = "monorepo/sdk"
	f.outputFlag = "services/"

	if (f.scopeFlag == "2" || f.scopeFlag == "3") && f.serviceName == "" {
		return fmt.Errorf(RedFormat, "missing service name, make sure to include '-service' flag")
	}
	return nil
}

func (f *flagParameter) validateServiceName() error {
	_, err := os.Stat(f.outputFlag + f.serviceName)
	if os.IsNotExist(err) {
		return fmt.Errorf(RedFormat, fmt.Sprintf(`Service "%s" is not exist in "%s" directory`, f.serviceName, f.outputFlag))
	}
	return nil
}

func (f *flagParameter) validateModuleName(moduleName string) (err error) {
	_, err = os.Stat(f.outputFlag + f.serviceName + "/internal/modules/" + moduleName)
	if os.IsNotExist(err) {
		fmt.Printf(RedFormat, fmt.Sprintf(`Module "%s" is not exist in service "%s"`, moduleName, f.serviceName))
		os.Exit(1)
	}
	return
}

func (f *flagParameter) getFullModuleChildDir(paths ...string) string {
	paths = append([]string{f.moduleName}, paths...)
	return strings.TrimPrefix(f.outputFlag+f.serviceName+"/internal/modules/"+strings.Join(paths, "/"), "/")
}

type configHeader struct {
	GoVersion     string
	Version       string
	Header        string `json:"-"`
	ServiceName   string
	PackagePrefix string
	ProtoSource   string
	OutputDir     string `json:"-"`
	Owner         string
	License       string
	Year          int
}

type config struct {
	IsMonorepo                                                         bool
	RestHandler, GRPCHandler, GraphQLHandler, FiberRestHandler         bool
	KafkaHandler, SchedulerHandler, RedisSubsHandler, TaskQueueHandler bool
	PostgresListenerHandler, RabbitMQHandler, IsWorkerActive           bool
	RedisDeps, SQLDeps, MongoDeps, SQLUseGORM, ArangoDeps              bool
	SQLDriver                                                          string
	WorkerPlugins                                                      []string
}

type serviceConfig struct {
	configHeader
	config
	Modules       []moduleConfig
	flag          *flagParameter    `json:"-"`
	workerPlugins map[string]plugin `json:"-"`
}

func (s *serviceConfig) getRootDir() (rootDir string) {
	if s.IsMonorepo || s.flag.initService {
		return s.flag.outputFlag + s.ServiceName + "/"
	}
	return
}

func (s *serviceConfig) parseDefaultHeader() {
	s.Header = fmt.Sprintf("Code generated by candi %s.", candi.Version)
	s.Version = candi.Version
	s.Year = time.Now().Year()
	s.GoVersion = getGoVersion()
}

func (s *config) checkWorkerActive() bool {
	s.IsWorkerActive = s.KafkaHandler ||
		s.SchedulerHandler ||
		s.RedisSubsHandler ||
		s.PostgresListenerHandler ||
		s.TaskQueueHandler ||
		s.RabbitMQHandler ||
		len(s.WorkerPlugins) > 0
	return s.IsWorkerActive
}
func (s *serviceConfig) disableAllHandler() {
	s.RestHandler = false
	s.GRPCHandler = false
	s.GraphQLHandler = false
	s.KafkaHandler = false
	s.SchedulerHandler = false
	s.RedisSubsHandler = false
	s.TaskQueueHandler = false
	s.PostgresListenerHandler = false
	s.RabbitMQHandler = false
}
func (s *serviceConfig) toJSONString() string {
	jsonSrvConfig, _ := json.Marshal(s)
	var configJSON bytes.Buffer
	json.Indent(&configJSON, jsonSrvConfig, "", "     ")
	return configJSON.String()
}

type moduleConfig struct {
	configHeader `json:"-"`
	config       `json:"-"`
	ModuleName   string
	Skip         bool `json:"-"`
}

func (m *moduleConfig) constructModuleWorkerActivation() (workerActivations []string) {
	if m.KafkaHandler {
		workerActivations = append(workerActivations, "types.Kafka")
	}
	if m.SchedulerHandler {
		workerActivations = append(workerActivations, "types.Scheduler")
	}
	if m.RedisSubsHandler {
		workerActivations = append(workerActivations, "types.RedisSubscriber")
	}
	if m.TaskQueueHandler {
		workerActivations = append(workerActivations, "types.TaskQueue")
	}
	if m.PostgresListenerHandler {
		workerActivations = append(workerActivations, "types.PostgresListener")
	}
	if m.RabbitMQHandler {
		workerActivations = append(workerActivations, "types.RabbitMQ")
	}
	return workerActivations
}

// FileStructure model
type FileStructure struct {
	TargetDir    string
	IsDir        bool
	FromTemplate bool
	DataSource   interface{}
	Source       string
	FileName     string
	Skip         bool
	SkipAll      bool
	SkipIfExist  bool
	Childs       []FileStructure
}

func (f *FileStructure) parseTemplate() (buff []byte) {
	if f.FromTemplate {
		if f.Source != "" {
			buff = loadTemplate(f.Source, f.DataSource)
		} else {
			lastDir := filepath.Dir(f.TargetDir)
			buff = defaultDataSource(lastDir[strings.LastIndex(lastDir, "/")+1:])
		}
	} else {
		buff = []byte(f.Source)
	}
	return
}

func (f *FileStructure) writeFile(targetPath string) error {
	fmt.Printf("creating %s...\n", targetPath+"/"+f.FileName)
	return os.WriteFile(targetPath+"/"+f.FileName, f.parseTemplate(), 0644)
}
