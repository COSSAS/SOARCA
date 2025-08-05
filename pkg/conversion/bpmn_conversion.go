package conversion

import (
	"encoding/xml"
	"errors"
	"github.com/google/uuid"
	"soarca/pkg/models/cacao"
)

type BpmnConverter struct {
	nodes []IDecodedBpmnNode
	tasks map[string]*DecodedBpmnTask
	start *DecodedBpmnStartEvent
	end   *DecodedBpmnEndEvent
}

func (converter *BpmnConverter) implement(playbook *cacao.Playbook) error {
	for _, node := range converter.nodes {
		if err := node.Implement(playbook); err != nil {
			return err
		}
	}
	return nil
}

// Basically everything in the playbook; after all is gathered we "Implement" all of these in the cacao playbook
type IDecodedBpmnNode interface {
	Implement(*cacao.Playbook) error
}
type DecodedBpmnTask struct {
	Kind     string
	Incoming *IDecodedBpmnNode
	Outgoing *IDecodedBpmnNode
	Uuid     uuid.UUID
}

func (e DecodedBpmnTask) Implement(*cacao.Playbook) error {
	return errors.New("Unimplemented: task")
}

type DecodedBpmnStartEvent struct {
	Outgoing *IDecodedBpmnNode
}

func (e DecodedBpmnStartEvent) Implement(*cacao.Playbook) error {
	return errors.New("Unimplemented: start")
}

type DecodedBpmnEndEvent struct {
	Outgoing *IDecodedBpmnNode
}

func (e DecodedBpmnEndEvent) Implement(*cacao.Playbook) error {
	return errors.New("Unimplemented: end")
}

func (converter BpmnConverter) Convert(input []byte) (*cacao.Playbook, error) {
	var file BpmnFile
	if err := xml.Unmarshal(input, &file); err != nil {
		return nil, err
	}
	if len(file.Processes) > 1 {
		return nil, errors.New("Unsupported: BPMN file with multiple processes")
	}
	if len(file.Processes) == 0 {
		return nil, errors.New("BPMN file does not have any processes")
	}
	converter.gather(file.Processes[0])
	playbook := cacao.NewPlaybook()
	converter.implement(playbook)
	return nil, errors.New("Unimplemented: further BPMN processing")
}
func NewBpmnConverter() BpmnConverter {
	return BpmnConverter{}
}

type BpmnFile struct {
	Processes []BpmnProcess `xml:"bpmn:definitions"`
}
type BpmnProcess struct {
	start_task *BpmnStartEvent
	end_task   *BpmnEndEvent
	flows      []BpmnFlow
	tasks      []BpmnTask
}

func (p *BpmnProcess) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	for {
		item, err := d.Token()
		if err != nil {
			return err
		}
		switch item_type := item.(type) {
		case xml.StartElement:
			switch item_type.Name.Local {
			case "bpmn:startEvent":
				err = d.DecodeElement(&p.start_task, &item_type)
				if err != nil {
					return err
				}
			case "bpmn:endEvent":
				err = d.DecodeElement(&p.end_task, &item_type)
				if err != nil {
					return err
				}
			case "bpmn:sequenceFlow":
				flow := new(BpmnFlow)
				err = d.DecodeElement(flow, &item_type)
				if err != nil {
					return err
				}
				p.flows = append(p.flows, *flow)
			case "bpmn:scriptTask":
				task := new(BpmnTask)
				task.Kind = "script"
				err = d.DecodeElement(task, &item_type)
				if err != nil {
					return err
				}
				p.tasks = append(p.tasks, *task)
			}
		}
	}
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
	Kind     string
	Id       string  `xml:"id,attr"`
	Incoming *string `xml:"bpmn:incoming"`
	Outgoing *string `xml:"bpmn:outgoing"`
}

type BpmnFlow struct {
	Id        string `xml:"id,attr"`
	SourceRef string `xml:"sourceRef,attr"`
	TargetRef string `xml:"targetRef,attr"`
}

func (converter *BpmnConverter) gather(process BpmnProcess) error {
	converter.start = &DecodedBpmnStartEvent{}
	converter.nodes = append(converter.nodes, converter.start)
	converter.end = &DecodedBpmnEndEvent{}
	converter.nodes = append(converter.nodes, converter.end)
	for _, task := range process.tasks {
		converted_task := new(DecodedBpmnTask)
		converted_task.Kind = task.Kind
		converted_task.Uuid = uuid.New()
		converter.tasks[task.Id] = converted_task
		converter.nodes = append(converter.nodes, converted_task)
	}
	return nil
}
