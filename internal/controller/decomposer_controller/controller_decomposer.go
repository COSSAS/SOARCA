package decomposer_controller

import (
	"soarca/pkg/core/decomposer"
)

type IController interface {
	NewDecomposer() decomposer.IDecomposer
}
