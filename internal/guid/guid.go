package guid

import "github.com/google/uuid"

/*
*
IGuid allows one to inject the uuid property and have it be available by mocking
*
*/
type IGuid interface {
	New() uuid.UUID
	String() string
}

type Guid struct {
	uuid.UUID
}

func (id *Guid) New() uuid.UUID {
	var uuid, _ = uuid.NewUUID()
	return uuid
}

func (id *Guid) String() string {
	return id.UUID.String()
}
