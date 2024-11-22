package database

import (
	playbookrepository "soarca/internal/database/playbook"
)

type IController interface {
	GetDatabaseInstance() playbookrepository.IPlaybookRepository
}
