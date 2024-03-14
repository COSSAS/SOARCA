package decomposer

import (
	"soarca/internal/decomposer"
)

type IController interface {
	NewDecomposer() decomposer.IDecomposer
}
