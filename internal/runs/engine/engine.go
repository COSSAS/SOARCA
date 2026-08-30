// Package engine builds the per-run workflow walker and owns all capability,
// executor and reporter wiring.
package engine

import (
	"reflect"
	"time"

	"soarca/internal/config"
	"soarca/internal/logger"
	"soarca/internal/storage"
	"soarca/internal/workflow"
	"soarca/pkg/core/capability"
	fincap "soarca/pkg/core/capability/fin"
	"soarca/pkg/core/capability/fin/queue"
	httpcap "soarca/pkg/core/capability/http"
	manualcap "soarca/pkg/core/capability/manual"
	"soarca/pkg/core/capability/manual/interaction"
	openc2cap "soarca/pkg/core/capability/openc2"
	pscap "soarca/pkg/core/capability/powershell"
	sshcap "soarca/pkg/core/capability/ssh"
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

var log *logger.Log

type empty struct{}

func init() {
	log = logger.Logger(reflect.TypeOf(empty{}).PkgPath(), logger.Info, "", logger.Json)
}

// Deps are the collaborators a walker needs, passed explicitly so the
// engine never depends on the runtime container.
type Deps struct {
	Interaction        interaction.ICapabilityInteraction
	Cache              downstreamreport.IDownStreamReporter
	FinQueue           *queue.Queue
	FinStore           storage.FinStore
	PlaybookStore      storage.PlaybookStore
	SkipCertValidation bool
	FinStaleAfter      time.Duration
	TheHive            config.TheHiveConfig
}

// Factory creates a workflow walker per playbook run.
type Factory struct {
	deps Deps
}

func New(deps Deps) *Factory {
	return &Factory{deps: deps}
}

// NewWalker builds a workflow walker for a single run.
func (f *Factory) NewWalker() workflow.Walker {
	sshCap := new(sshcap.SshCapability)
	capabilities := map[string]capability.ICapability{sshCap.GetType(): sshCap}

	httpUtil := new(httputil.HttpRequest)
	httpUtil.SkipCertificateValidation(f.deps.SkipCertValidation)
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
		StaleAfter: f.deps.FinStaleAfter,
	}))

	pbExec := pbactionexec.New(f.NewWalker, f.deps.PlaybookStore, report, soarcaTime)
	stixCmp := stixcmp.New()
	condExec := condexec.New(stixCmp, report, soarcaTime)
	guidGen := new(guid.Guid)

	return workflow.New(actionExec, pbExec, condExec, guidGen, report, soarcaTime, theHiveCaseManager)
}

// initializeTheHiveReporting sets up The Hive integration if configured.
func (f *Factory) initializeTheHiveReporting() (downstreamreport.IDownStreamReporter, cases.ICasesManager) {
	cfg := f.deps.TheHive
	if !cfg.Activate {
		return nil, nil
	}

	log.Info("Initializing The Hive reporting integration")

	if len(cfg.APIBaseURL) < 1 || len(cfg.APIToken) < 1 {
		log.Warning("Could not initialize The Hive reporting integration. Check environment variables.")
		return nil, nil
	}

	log.Infof("Creating The Hive connector with API base URL: %s", cfg.APIBaseURL)
	conn := thehiveconnector.NewConnector(cfg.APIBaseURL, cfg.APIToken, cfg.AllowInsecure)

	if cfg.EnableCaseManager {
		log.Info("Enabling The Hive case manager")
		caseMgr := thehivecases.NewCaseManager(conn)
		return caseMgr, caseMgr
	}

	if cfg.EnableReporter {
		log.Info("Enabling The Hive reporter")
		rep := thehivereport.NewReporter(conn)
		return rep, nil
	}

	return nil, nil
}
