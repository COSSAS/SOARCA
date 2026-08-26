// Package finrepository is the persistence layer for Fin registrations
// (pkg/models/fin.Record). Unlike the in-memory job queue
// (pkg/core/capability/fin/queue), Fin registrations are persisted so a Fin
// process does not need to re-register every time SOARCA restarts or is
// updated.
package finrepository

import (
	"errors"
	"reflect"
	"time"

	"soarca/internal/database"
	"soarca/internal/logger"
	"soarca/pkg/models/fin"
)

type Empty struct{}

var component = reflect.TypeOf(Empty{}).PkgPath()
var log *logger.Log

func init() {
	log = logger.Logger(component, logger.Info, "", logger.Json)
}

// IFinRepository is the persistence contract for Fin registrations.
type IFinRepository interface {
	// Register persists a new Fin registration. Returns an error if
	// record.FinId already exists.
	Register(record fin.Record) error
	// Get looks up a Fin registration by its FinId. Returns
	// fin.ErrFinNotFound if none exists.
	Get(finId string) (fin.Record, error)
	// FindByTokenHash looks up the Fin registration whose FinTokenHash
	// matches tokenHash - used to authenticate poll/result/status-ping/
	// unregister calls, which present a fin_token rather than a fin_id.
	// Returns fin.ErrFinTokenInvalid if no registration matches.
	FindByTokenHash(tokenHash string) (fin.Record, error)
	// List returns every currently-registered Fin.
	List() ([]fin.Record, error)
	// Touch updates a Fin's LastSeen timestamp - called on every
	// successful poll, whether or not it returned a job, since polling
	// itself is this protocol's liveness signal.
	Touch(finId string, lastSeen time.Time) error
	// Unregister removes a Fin's registration.
	Unregister(finId string) error
}

// FinRepository is the database.Database-backed IFinRepository
// implementation (see internal/database/mongodb for the concrete
// database.Database this is normally constructed against).
type FinRepository struct {
	db database.Database
}

func SetupFinRepository(db database.Database) *FinRepository {
	return &FinRepository{db: db}
}

var _ IFinRepository = (*FinRepository)(nil)

func (finRepo *FinRepository) Register(record fin.Record) error {
	if record.RegisteredAt.IsZero() {
		record.RegisteredAt = time.Now()
	}
	if record.LastSeen.IsZero() {
		record.LastSeen = record.RegisteredAt
	}
	err := finRepo.db.Create(record)
	if err != nil {
		log.Error("failed to register fin ", record.FinId, ": ", err)
	}
	return err
}

func (finRepo *FinRepository) Get(finId string) (fin.Record, error) {
	returnedObject, err := finRepo.db.Read(finId)
	if err != nil {
		return fin.Record{}, fin.ErrFinNotFound{FinId: finId}
	}
	record, ok := returnedObject.(fin.Record)
	if !ok {
		return fin.Record{}, errors.New("could not cast lookup object to fin.Record type")
	}
	return record, nil
}

func (finRepo *FinRepository) FindByTokenHash(tokenHash string) (fin.Record, error) {
	records, err := finRepo.db.Find(map[string]string{"fin_token_hash": tokenHash})
	if err != nil {
		return fin.Record{}, err
	}
	if len(records) == 0 {
		return fin.Record{}, fin.ErrFinTokenInvalid{}
	}
	record, ok := records[0].(fin.Record)
	if !ok {
		return fin.Record{}, errors.New("could not cast lookup object to fin.Record type")
	}
	return record, nil
}

func (finRepo *FinRepository) List() ([]fin.Record, error) {
	records, err := finRepo.db.Find(map[string]string{})
	if err != nil {
		return nil, err
	}
	returnRecords := make([]fin.Record, 0, len(records))
	for _, record := range records {
		record, ok := record.(fin.Record)
		if !ok {
			return nil, errors.New("could not cast lookup object to fin.Record type")
		}
		returnRecords = append(returnRecords, record)
	}
	return returnRecords, nil
}

func (finRepo *FinRepository) Touch(finId string, lastSeen time.Time) error {
	return finRepo.db.Update(finId, map[string]any{"last_seen": lastSeen})
}

func (finRepo *FinRepository) Unregister(finId string) error {
	return finRepo.db.Delete(finId)
}
