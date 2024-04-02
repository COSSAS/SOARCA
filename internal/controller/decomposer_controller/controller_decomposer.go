package decomposer_controller

import (
	"soarca/internal/decomposer"
)

type IController interface {
	NewDecomposer() decomposer.IDecomposer
}
