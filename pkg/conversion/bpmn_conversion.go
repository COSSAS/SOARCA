package conversion

import (
	"encoding/xml"
	"errors"
	"fmt"
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
	Id        string `xml:"id,attr"`
	SourceRef string `xml:"sourceRef,attr"`
	TargetRef string `xml:"targetRef,attr"`
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
	playbook.Workflow = make(cacao.Workflow)
	converter.implement(definitions.Processes[0], playbook)
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
	start_task *BpmnStartEvent
	end_task   *BpmnEndEvent
	flows      []BpmnFlow
	tasks      []BpmnTask
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
				err = d.DecodeElement(&p.end_task, &item_type)
				if err != nil {
					return err
				}
			case "sequenceFlow":
				flow := new(BpmnFlow)
				err = d.DecodeElement(flow, &item_type)
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
	log.Info("Implementing end task ", process.end_task.Id)
	process.end_task.implement(playbook, converter)
	for _, task := range process.tasks {
		log.Info("Implementing task ", task.Name)
		if err := task.implement(playbook, converter); err != nil {
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
	step := cacao.Step{Type: "action", Name: task.Name}
	playbook.Workflow[name] = step
	return nil
}
func (flow BpmnFlow) implement(playbook *cacao.Playbook, converter *BpmnConverter) error {
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
	entry.OnCompletion = target_name
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
