package interaction

import (
	"soarca/pkg/core/capability"
	"soarca/pkg/models/cacao"
	"soarca/pkg/models/execution"
)

type InteractionCommand struct {
	Metadata execution.Metadata
	Context  capability.Context
}

type InteractionResponse struct {
	ResponseError error
	Variables     cacao.Variables
}

type ICapabilityInteraction interface {
	Queue(InteractionCommand, chan InteractionResponse) error
}
