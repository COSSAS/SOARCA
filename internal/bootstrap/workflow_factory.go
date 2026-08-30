package bootstrap

import (
	"reflect"

	"soarca/internal/config"
	"soarca/internal/logger"
	"soarca/internal/storage"
	"soarca/pkg/core/capability"
	fincap "soarca/pkg/core/capability/fin"
	"soarca/pkg/core/capability/fin/queue"
	httpcap "soarca/pkg/core/capability/http"
	manualcap "soarca/pkg/core/capability/manual"
	"soarca/pkg/core/capability/manual/interaction"
	openc2cap "soarca/pkg/core/capability/openc2"
	pscap "soarca/pkg/core/capability/powershell"
	sshcap "soarca/pkg/core/capability/ssh"
	"soarca/pkg/core/decomposer"
	actionexec "soarca/pkg/core/executors/action"
	condexec "soarca/pkg/core/executors/condition"
	pbactionexec "soarca/pkg/core/executors/playbook_action"
	"soarca/pkg/extensions/soarca/assignment"
	thehivecases "soarca/pkg/integration/thehive/cases"
	thehiveconnector "soarca/pkg/integration/thehive/common/connector"
	thehivereport "soarca/pkg/integration/thehive/reporter"
	"soarca/pkg/reporting/cases"
	"soarca/pkg/reporting/reporter"
	downstreamreport "soarca/pkg/reporting/reporter/downstream_reporter"
	"soarca/pkg/utils/guid"
	httputil "soarca/pkg/utils/http"
	stixcmp "soarca/pkg/utils/stix/expression/comparison"
	timeutil "soarca/pkg/utils/time"
)

var wflog *logger.Log

type wfEmpty struct{}

func init() {
	wflog = logger.Logger(reflect.TypeOf(wfEmpty{}).PkgPath(), logger.Info, "", logger.Json)
}

// EngineDeps are the collaborators a decomposer needs. They are passed
// explicitly so the engine never depends on the runtime container.
type EngineDeps struct {
	Interaction   interaction.ICapabilityInteraction
	Cache         downstreamreport.IDownStreamReporter
	FinQueue      *queue.Queue
	FinStore      storage.FinStore
	PlaybookStore storage.PlaybookStore
	Config        config.Config
}

// WorkflowFactory creates decomposers for playbook execution.
// It owns all capability wiring (SSH, HTTP, OpenC2, PowerShell, Manual, FIN),
// the reporter chain, and optional TheHive integration.
type WorkflowFactory struct {
	deps EngineDeps
}

func newWorkflowFactory(deps EngineDeps) *WorkflowFactory {
	return &WorkflowFactory{deps: deps}
}

// NewDecomposer implements decomposer_controller.IController.
// Called by the execution runtime once per playbook execution.
func (f *WorkflowFactory) NewDecomposer() decomposer.IDecomposer {
	sshCap := new(sshcap.SshCapability)
	capabilities := map[string]capability.ICapability{sshCap.GetType(): sshCap}

	httpUtil := new(httputil.HttpRequest)
	httpUtil.SkipCertificateValidation(f.deps.Config.HTTP.SkipCertValidation)
	httpCap := httpcap.New(httpUtil)
	capabilities[httpCap.GetType()] = httpCap

	openc2Cap := openc2cap.New(httpUtil)
	capabilities[openc2Cap.GetType()] = openc2Cap

	powershellCap := pscap.New()
	capabilities[powershellCap.GetType()] = powershellCap

	man := manualcap.New(f.deps.Interaction)
	capabilities[man.GetType()] = &man

	report := reporter.New([]downstreamreport.IDownStreamReporter{})
	downstreamReporters := []downstreamreport.IDownStreamReporter{f.deps.Cache}

	thehiveReporter, theHiveCaseManager := f.initializeTheHiveReporting()
	if thehiveReporter != nil {
		downstreamReporters = append(downstreamReporters, thehiveReporter)
	}

	report.RegisterReporters(downstreamReporters)

	soarcaTime := new(timeutil.Time)
	assignmentExt := assignment.New()
	actionExec := actionexec.New(capabilities, report, soarcaTime, assignmentExt)

	actionExec.SetFinFallback(fincap.New(fincap.Dependencies{
		Queue:      f.deps.FinQueue,
		GUID:       new(guid.Guid),
		Store:      f.deps.FinStore,
		Time:       soarcaTime,
		StaleAfter: f.deps.Config.Fin.StaleAfter,
	}))

	pbExec := pbactionexec.New(f, f.deps.PlaybookStore, report, soarcaTime)
	stixCmp := stixcmp.New()
	condExec := condexec.New(stixCmp, report, soarcaTime)
	guidGen := new(guid.Guid)
	decomp := decomposer.New(actionExec, pbExec, condExec, guidGen, report, soarcaTime)

	if theHiveCaseManager != nil {
		decomp.SetCaseManager(theHiveCaseManager)
	}

	return decomp
}

// initializeTheHiveReporting sets up The Hive integration if configured.
func (f *WorkflowFactory) initializeTheHiveReporting() (downstreamreport.IDownStreamReporter, cases.ICasesManager) {
	cfg := f.deps.Config
	if !cfg.TheHive.Activate {
		return nil, nil
	}

	wflog.Info("Initializing The Hive reporting integration")

	if len(cfg.TheHive.APIBaseURL) < 1 || len(cfg.TheHive.APIToken) < 1 {
		wflog.Warning("Could not initialize The Hive reporting integration. Check environment variables.")
		return nil, nil
	}

	wflog.Infof("Creating The Hive connector with API base URL: %s", cfg.TheHive.APIBaseURL)
	conn := thehiveconnector.NewConnector(cfg.TheHive.APIBaseURL, cfg.TheHive.APIToken, cfg.TheHive.AllowInsecure)

	if cfg.TheHive.EnableCaseManager {
		wflog.Info("Enabling The Hive case manager")
		caseMgr := thehivecases.NewCaseManager(conn)
		return caseMgr, caseMgr
	}

	if cfg.TheHive.EnableReporter {
		wflog.Info("Enabling The Hive reporter")
		rep := thehivereport.NewReporter(conn)
		return rep, nil
	}

	return nil, nil
}
