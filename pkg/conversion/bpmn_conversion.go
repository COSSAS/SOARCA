package conversion

import (
	"encoding/xml"
	"errors"
	"fmt"
	"slices"
	"soarca/internal/logger"
	"soarca/pkg/models/cacao"

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

func (converter BpmnConverter) Convert(input []byte) (*cacao.Playbook, error) {
	converter.translation = make(map[string]string)
	var definitions BpmnDefinitions
	if err := xml.Unmarshal(input, &definitions); err != nil {
		return nil, err
	}
	if len(definitions.Processes) > 1 {
		return nil, errors.New("Unsupported: BPMN file with multiple processes")
	}
	if len(definitions.Processes) == 0 {
		return nil, errors.New("BPMN file does not have any processes")
	}
	playbook := cacao.NewPlaybook()
	playbook.SpecVersion = cacao.CACAO_VERSION_2
	playbook.Type = "playbook"
	playbook.ID = fmt.Sprintf("playbook--%s", uuid.New())
	playbook.CreatedBy = fmt.Sprintf("identity--%s", uuid.New())
	soarca_name := fmt.Sprintf("soarca--%s", uuid.New())
	soarca_manual_name := fmt.Sprintf("soarca--%s", uuid.New())
	converter.translation["soarca"] = soarca_name
	converter.translation["soarca-manual"] = soarca_manual_name
	playbook.AgentDefinitions = cacao.NewAgentTargets(
		cacao.AgentTarget{
			ID:   soarca_name,
			Type: "soarca",
			Name: "soarca-playbook"},
		cacao.AgentTarget{
			ID:          soarca_manual_name,
			Type:        "soarca-manual",
			Name:        "soarca-manual",
			Description: "SOARCAs manual command handler"})
	playbook.TargetDefinitions = cacao.NewAgentTargets(
		cacao.AgentTarget{
			ID:   fmt.Sprintf("individual--%s", uuid.New()),
			Type: "individual",
			Name: "CHANGE THIS"})
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

func (p *BpmnProcess) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	start_name := start.Name
	for {
		item, err := d.Token()
		if err != nil {
			return err
		}
		switch item_type := item.(type) {
		case xml.StartElement:
			switch item_type.Name.Local {
			case "startEvent":
				err = d.DecodeElement(&p.start_task, &item_type)
				if err != nil {
					return err
				}
			case "endEvent":
				end_task := BpmnEndEvent{}
				err = d.DecodeElement(&end_task, &item_type)
				p.end_tasks = append(p.end_tasks, end_task)
				if err != nil {
					return err
				}
			case "sequenceFlow":
				flow := new(BpmnFlow)
				err = d.DecodeElement(flow, &item_type)
				flow.IsAssociation = false
				if err != nil {
					return err
				}
				p.flows = append(p.flows, *flow)
			case "association":
				flow := new(BpmnFlow)
				err = d.DecodeElement(flow, &item_type)
				flow.IsAssociation = true
				if err != nil {
					return err
				}
				p.flows = append(p.flows, *flow)
			case "scriptTask":
				task := new(BpmnTask)
				task.Kind = "script"
				err = d.DecodeElement(task, &item_type)
				if err != nil {
					return err
				}
				p.tasks = append(p.tasks, *task)
			case "task":
				task := new(BpmnTask)
				task.Kind = "task"
				err = d.DecodeElement(task, &item_type)
				if err != nil {
					return err
				}
				p.tasks = append(p.tasks, *task)
			case "serviceTask":
				task := new(BpmnTask)
				task.Kind = "service"
				err = d.DecodeElement(task, &item_type)
				if err != nil {
					return err
				}
				p.tasks = append(p.tasks, *task)
			case "sendTask":
				task := new(BpmnTask)
				task.Kind = "send"
				err = d.DecodeElement(task, &item_type)
				if err != nil {
					return err
				}
				p.tasks = append(p.tasks, *task)
			case "userTask":
				task := new(BpmnTask)
				task.Kind = "user"
				err = d.DecodeElement(task, &item_type)
				if err != nil {
					return err
				}
				p.tasks = append(p.tasks, *task)
			case "businessRuleTask":
				task := new(BpmnTask)
				task.Kind = "business rule"
				err = d.DecodeElement(task, &item_type)
				if err != nil {
					return err
				}
				p.tasks = append(p.tasks, *task)
			case "exclusiveGateway":
				gateway := new(BpmnGateway)
				err = d.DecodeElement(gateway, &item_type)
				gateway.Kind = GatewayKindExclusive
				if err != nil {
					return err
				}
				p.gateways = append(p.gateways, *gateway)
			case "parallelGateway":
				gateway := new(BpmnGateway)
				err = d.DecodeElement(gateway, &item_type)
				gateway.Kind = GatewayKindParallel
				if err != nil {
					return err
				}
				p.gateways = append(p.gateways, *gateway)
			case "textAnnotation":
				annotation := new(BpmnAnnotation)
				err = d.DecodeElement(annotation, &item_type)
				if err != nil {
					return err
				}
				p.annotations = append(p.annotations, *annotation)
			case "intermediateThrowEvent", "intermediateCatchEvent":
				return fmt.Errorf("Throw/catch mechanism is currently not implemented in SOARCA")
			default:
				return fmt.Errorf("Unsupported element: %s", item_type.Name.Local)
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
	process.start_task.implement(playbook, converter)
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
	name := fmt.Sprintf("action--%s", uuid.New())
	converter.translation[task.Id] = name
	step := cacao.Step{Type: "action", Name: task.Name, Commands: make([]cacao.Command, 0)}
	step.Commands = append(step.Commands, cacao.Command{Type: "manual", Command: task.Name})
	step.Agent = converter.translation["soarca"]
	playbook.Workflow[name] = step
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
		return fmt.Errorf("Could not translate source of flow: %s", flow.SourceRef)
	}
	source := playbook.Workflow[source_name]
	target_index := slices.IndexFunc(converter.process.annotations, func(annot BpmnAnnotation) bool { return annot.Id == flow.TargetRef })
	if target_index < 0 {
		return fmt.Errorf("Could not find text annotation %s", flow.TargetRef)
	}
	target := converter.process.annotations[target_index]
	source.Condition = target.Text
	playbook.Workflow[source_name] = source
	return nil
}
func (flow BpmnFlow) implement_flow(playbook *cacao.Playbook, converter *BpmnConverter) error {
	source_name, ok := converter.translation[flow.SourceRef]
	if !ok {
		return fmt.Errorf("Could not translate source of flow: %s", flow.SourceRef)
	}
	target_name, ok := converter.translation[flow.TargetRef]
	if !ok {
		return fmt.Errorf("Could not translate target of flow: %s", flow.TargetRef)
	}
	log.Infof("Flow from %s(%s) to %s(%s)", source_name, flow.SourceRef, target_name, flow.TargetRef)
	entry, ok := playbook.Workflow[source_name]
	if !ok {
		return fmt.Errorf("Could not get source of flow: %s", source_name)
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
				return fmt.Errorf("Branch out of exclusive gateway with more than two branches: not supported")
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

func (end_event BpmnEndEvent) implement(playbook *cacao.Playbook, converter *BpmnConverter) error {
	name := fmt.Sprintf("end--%s", uuid.New())
	converter.translation[end_event.Id] = name
	step := cacao.Step{Type: "end"}
	playbook.Workflow[name] = step
	playbook.WorkflowException = name
	return nil
}
func (start_event BpmnStartEvent) implement(playbook *cacao.Playbook, converter *BpmnConverter) error {
	name := fmt.Sprintf("start--%s", uuid.New())
	converter.translation[start_event.Id] = name
	step := cacao.Step{Type: "start"}
	playbook.Workflow[name] = step
	playbook.WorkflowStart = name
	return nil
}
func (gateway BpmnGateway) implement(playbook *cacao.Playbook, converter *BpmnConverter) error {
	switch gateway.Kind {
	case GatewayKindExclusive:
		return gateway.implement_exclusive(playbook, converter)
	case GatewayKindParallel:
		return gateway.implement_parallel(playbook, converter)
	}
	return nil
}
func (gateway BpmnGateway) implement_exclusive(playbook *cacao.Playbook, converter *BpmnConverter) error {
	condition := cacao.Step{
		Type:      cacao.StepTypeIfCondition,
		Condition: gateway.Name,
	}
	name := fmt.Sprintf("if-condition--%s", uuid.New())
	converter.translation[gateway.Id] = name
	playbook.Workflow[name] = condition
	return nil
}
func (gateway BpmnGateway) implement_parallel(playbook *cacao.Playbook, converter *BpmnConverter) error {
	condition := cacao.Step{
		Type:      cacao.StepTypeParallel,
		Condition: gateway.Name,
	}
	name := fmt.Sprintf("parallel--%s", uuid.New())
	converter.translation[gateway.Id] = name
	playbook.Workflow[name] = condition
	return nil
}
