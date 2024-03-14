package database

import (
	playbookrepository "soarca/database/playbook"
)

type IController interface {
	GetDatabaseInstance() playbookrepository.IPlaybookRepository
}
