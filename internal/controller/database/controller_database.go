package database

import "soarca/internal/storage"

type IController interface {
	GetPlaybookStore() storage.PlaybookStore
}
