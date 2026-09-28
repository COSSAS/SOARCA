package conversion

import (
	"encoding/xml"
	"errors"
	"fmt"
	"path"
	"slices"
	"soarca/internal/logger"
	"soarca/pkg/models/cacao"
	util "soarca/pkg/utils/conversion"
	"strings"

	"github.com/google/uuid"
)

var log *logger.Log

func init() {
	log = logger.Logger("PBMN", logger.Info, "", logger.Json)
}

type BpmnConverter struct {
	translation map[string]string
	process     *BpmnProcess
}

/* This structure is somewhat unfortunate; the BPMN process definition is a non-homogeneous list
** of different playbook componenents, including tasks, the arrows/"flows" between tasks, and
** gateways (which become if/else constructs in cacao).
** We could put all of their associated fields and attributes in one process-item type, but this would shift
** the burden of finding out what fields belong to what kind of element to the developer. This is quite error-prone,
** and also would be hard to maintain.
** The current implementation involves this collection of types for different kinds of BPMN process elements, with a
** custom XML decoding step to collect them.
 */
type BpmnStartEvent struct {
	Id       string `xml:"id,attr"`
	Outgoing string `xml:"bpmn:outgoing"`
}
type BpmnEndEvent struct {
	Id       string `xml:"id,attr"`
	Incoming string `xml:"bpmn:incoming"`
}
type BpmnTask struct {
	Kind string
	Id   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

type BpmnFlow struct {
	Id            string `xml:"id,attr"`
	SourceRef     string `xml:"sourceRef,attr"`
	TargetRef     string `xml:"targetRef,attr"`
	Name          string `xml:"name,attr"`
	IsAssociation bool
}

type BpmnGatewayKind int

const (
	GatewayKindExclusive BpmnGatewayKind = iota
	GatewayKindParallel
)

type BpmnGateway struct {
	Id   string `xml:"id,attr"`
	Name string `xml:"name,attr"`
	Kind BpmnGatewayKind
}

type BpmnAnnotation struct {
	Id   string `xml:"id,attr"`
	Text string `xml:"text"`
}

func clean_filename(filename string) string {
	base := path.Base(filename)
	ext := path.Ext(base)
	return strings.TrimSuffix(base, ext)
}

func (converter BpmnConverter) Convert(input []byte, filename string) (*cacao.Playbook, error) {
	converter.translation = make(map[string]string)
	var definitions BpmnDefinitions
	if err := xml.Unmarshal(input, &definitions); err != nil {
		return nil, err
	}
	if len(definitions.Processes) > 1 {
		return nil, errors.New("unsupported: BPMN file with multiple processes")
	}
	if len(definitions.Processes) == 0 {
		return nil, errors.New("BPMN file does not have any processes")
	}
	playbook, soarca_name, soarca_manual_name := util.NewSoarcaPlaybook(clean_filename(filename), []string{"notification"})
	playbook.Description = fmt.Sprintf("CACAO playbook converted from %s", filename)
	converter.translation["soarca"] = soarca_name
	converter.translation["soarca-manual"] = soarca_manual_name
	playbook.TargetDefinitions = cacao.NewAgentTargets(
		cacao.AgentTarget{
			ID:   fmt.Sprintf("individual--%s", uuid.New()),
			Type: "individual",
			Name: ""})
	playbook.Workflow = make(cacao.Workflow)
	converter.process = &definitions.Processes[0]
	if err := converter.implement(definitions.Processes[0], playbook); err != nil {
		return nil, err
	}
	return playbook, nil
}
func NewBpmnConverter() BpmnConverter {
	return BpmnConverter{}
}

type BpmnFile struct {
	Definition BpmnDefinitions `xml:"definitions"`
}
type BpmnDefinitions struct {
	Processes []BpmnProcess `xml:"process"`
}
type BpmnProcess struct {
	start_task  *BpmnStartEvent
	end_tasks   []BpmnEndEvent
	flows       []BpmnFlow
	tasks       []BpmnTask
	gateways    []BpmnGateway
	annotations []BpmnAnnotation
}

func (process *BpmnProcess) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	start_name := start.Name
	for {
		item, err := decoder.Token()
		if err != nil {
			return err
		}
		switch item_type := item.(type) {
		case xml.StartElement:
			switch item_type.Name.Local {
			case "startEvent":
				err = decoder.DecodeElement(&process.start_task, &item_type)
				if err != nil {
					return err
				}
			case "endEvent":
				end_task := BpmnEndEvent{}
				err = decoder.DecodeElement(&end_task, &item_type)
				process.end_tasks = append(process.end_tasks, end_task)
				if err != nil {
					return err
				}
			case "sequenceFlow":
				flow := new(BpmnFlow)
				err = decoder.DecodeElement(flow, &item_type)
				flow.IsAssociation = false
				if err != nil {
					return err
				}
				process.flows = append(process.flows, *flow)
			case "association":
				flow := new(BpmnFlow)
				err = decoder.DecodeElement(flow, &item_type)
				flow.IsAssociation = true
				if err != nil {
					return err
				}
				process.flows = append(process.flows, *flow)
			case "scriptTask":
				task := new(BpmnTask)
				task.Kind = "script"
				err = decoder.DecodeElement(task, &item_type)
				if err != nil {
					return err
				}
				process.tasks = append(process.tasks, *task)
			case "task":
				task := new(BpmnTask)
				task.Kind = "task"
				err = decoder.DecodeElement(task, &item_type)
				if err != nil {
					return err
				}
				process.tasks = append(process.tasks, *task)
			case "serviceTask":
				task := new(BpmnTask)
				task.Kind = "service"
				err = decoder.DecodeElement(task, &item_type)
				if err != nil {
					return err
				}
				process.tasks = append(process.tasks, *task)
			case "sendTask":
				task := new(BpmnTask)
				task.Kind = "send"
				err = decoder.DecodeElement(task, &item_type)
				if err != nil {
					return err
				}
				process.tasks = append(process.tasks, *task)
			case "userTask":
				task := new(BpmnTask)
				task.Kind = "user"
				err = decoder.DecodeElement(task, &item_type)
				if err != nil {
					return err
				}
				process.tasks = append(process.tasks, *task)
			case "businessRuleTask":
				task := new(BpmnTask)
				task.Kind = "business rule"
				err = decoder.DecodeElement(task, &item_type)
				if err != nil {
					return err
				}
				process.tasks = append(process.tasks, *task)
			case "exclusiveGateway":
				gateway := new(BpmnGateway)
				err = decoder.DecodeElement(gateway, &item_type)
				gateway.Kind = GatewayKindExclusive
				if err != nil {
					return err
				}
				process.gateways = append(process.gateways, *gateway)
			case "parallelGateway":
				gateway := new(BpmnGateway)
				err = decoder.DecodeElement(gateway, &item_type)
				gateway.Kind = GatewayKindParallel
				if err != nil {
					return err
				}
				process.gateways = append(process.gateways, *gateway)
			case "textAnnotation":
				annotation := new(BpmnAnnotation)
				err = decoder.DecodeElement(annotation, &item_type)
				if err != nil {
					return err
				}
				process.annotations = append(process.annotations, *annotation)
			case "intermediateThrowEvent", "intermediateCatchEvent":
				return fmt.Errorf("throw/catch mechanism is currently not implemented in SOARCA")
			default:
				return fmt.Errorf("unsupported element: %s", item_type.Name.Local)
			}
		case xml.EndElement:
			if item_type.Name == start_name {
				return nil
			}
		}
	}
}

func (converter *BpmnConverter) implement(process BpmnProcess, playbook *cacao.Playbook) error {
	log.Info("Implementing start task ", process.start_task.Id)
	if err := process.start_task.implement(playbook, converter); err != nil {
		return err
	}
	for _, end := range process.end_tasks {
		log.Info("Implementing end ", end.Id)
		if err := end.implement(playbook, converter); err != nil {
			return err
		}
	}
	log.Infof("Implementing %d tasks, %d gateways, and %d flows", len(process.tasks), len(process.gateways), len(process.flows))
	for _, task := range process.tasks {
		log.Info("Implementing task ", task.Name)
		if err := task.implement(playbook, converter); err != nil {
			return err
		}
	}
	for _, gateway := range process.gateways {
		log.Info("Implementing gateway ", gateway.Name)
		if err := gateway.implement(playbook, converter); err != nil {
			return err
		}
	}
	for _, flow := range process.flows {
		log.Info("Implementing flow ", flow.Id)
		if err := flow.implement(playbook, converter); err != nil {
			return err
		}
	}
	return nil
}

func (task BpmnTask) implement(playbook *cacao.Playbook, converter *BpmnConverter) error {
	step_id := fmt.Sprintf("action--%s", uuid.New())
	converter.translation[task.Id] = step_id
	step := cacao.Step{Type: "action", Name: task.Name, Commands: []cacao.Command{{Type: "manual", Command: task.Name}}}
	step.Agent = converter.translation["soarca"]
	playbook.Workflow[step_id] = step
	return nil
}
func create_step(type_ string, step_name string, playbook *cacao.Playbook, converter *BpmnConverter) string {
	step_id := fmt.Sprintf("%s--%s", type_, uuid.New())
	converter.translation[step_name] = step_id
	step := cacao.Step{Type: type_}
	playbook.Workflow[step_id] = step
	return step_id
}
func (end_event BpmnEndEvent) implement(playbook *cacao.Playbook, converter *BpmnConverter) error {
	playbook.WorkflowException = create_step("end", end_event.Id, playbook, converter)
	return nil
}
func (start_event BpmnStartEvent) implement(playbook *cacao.Playbook, converter *BpmnConverter) error {
	playbook.WorkflowStart = create_step("start", start_event.Id, playbook, converter)
	return nil
}

func (flow BpmnFlow) implement(playbook *cacao.Playbook, converter *BpmnConverter) error {
	if flow.IsAssociation {
		return flow.implement_association(playbook, converter)
	} else {
		return flow.implement_flow(playbook, converter)
	}
}

func (flow BpmnFlow) implement_association(playbook *cacao.Playbook, converter *BpmnConverter) error {
	source_name, ok := converter.translation[flow.SourceRef]
	if !ok {
		return fmt.Errorf("could not translate source of flow: %s", flow.SourceRef)
	}
	source := playbook.Workflow[source_name]
	target_index := slices.IndexFunc(converter.process.annotations, func(annot BpmnAnnotation) bool { return annot.Id == flow.TargetRef })
	if target_index < 0 {
		return fmt.Errorf("could not find text annotation %s", flow.TargetRef)
	}
	target := converter.process.annotations[target_index]
	source.Condition = target.Text
	playbook.Workflow[source_name] = source
	return nil
}
func (flow BpmnFlow) implement_flow(playbook *cacao.Playbook, converter *BpmnConverter) error {
	source_name, ok := converter.translation[flow.SourceRef]
	if !ok {
		return fmt.Errorf("could not translate source of flow: %s", flow.SourceRef)
	}
	target_name, ok := converter.translation[flow.TargetRef]
	if !ok {
		return fmt.Errorf("could not translate target of flow: %s", flow.TargetRef)
	}
	log.Infof("Flow from %s(%s) to %s(%s)", source_name, flow.SourceRef, target_name, flow.TargetRef)
	entry, ok := playbook.Workflow[source_name]
	if !ok {
		return fmt.Errorf("could not get source of flow: %s", source_name)
	}
	switch entry.Type {
	case cacao.StepTypeIfCondition:
		switch flow.Name {
		case "Yes", "yes":
			entry.OnTrue = target_name
		case "No", "no":
			entry.OnFalse = target_name
		default:
			log.Infof("Unknown flow name %s out of if-condition: picking empty branch", flow.Name)
			if entry.OnTrue == "" {
				entry.OnTrue = target_name
			} else if entry.OnFalse == "" {
				entry.OnTrue = target_name
			} else {
				return fmt.Errorf("branch out of exclusive gateway with more than two branches: not supported")
			}
		}
	case cacao.StepTypeParallel:
		entry.NextSteps = append(entry.NextSteps, target_name)
	default:
		entry.OnCompletion = target_name

	}
	playbook.Workflow[source_name] = entry
	return nil
}

func (gateway BpmnGateway) implement(playbook *cacao.Playbook, converter *BpmnConverter) error {
	switch gateway.Kind {
	case GatewayKindExclusive:
		return gateway.implement_gateway("if-condition", cacao.StepTypeIfCondition, playbook, converter)
	case GatewayKindParallel:
		return gateway.implement_gateway("parallel", cacao.StepTypeParallel, playbook, converter)
	}
	return nil
}
func (gateway BpmnGateway) implement_gateway(step_id_prefix string, condition_type string, playbook *cacao.Playbook, converter *BpmnConverter) error {
	condition := cacao.Step{
		Type:      condition_type,
		Condition: gateway.Name,
	}
	step_id := fmt.Sprintf("%s--%s", step_id_prefix, uuid.New())
	converter.translation[gateway.Id] = step_id
	playbook.Workflow[step_id] = condition
	return nil
}
