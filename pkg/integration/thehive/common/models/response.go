package models

// Case POST response
type CaseResponse struct {
	ID                  string         `json:"_id,omitempty"`
	Type                string         `json:"_type,omitempty"`
	CreatedBy           string         `json:"_createdBy,omitempty"`
	UpdatedBy           string         `json:"_updatedBy,omitempty"`
	CreatedAt           int64          `json:"_createdAt,omitempty"`
	UpdatedAt           int64          `json:"_updatedAt,omitempty"`
	Number              int            `json:"number,omitempty"`
	Title               string         `json:"title,omitempty"`
	Description         string         `json:"description,omitempty"`
	Severity            int            `json:"severity,omitempty"`
	SeverityLabel       string         `json:"severityLabel,omitempty"`
	StartDate           int64          `json:"startDate,omitempty"`
	EndDate             int64          `json:"endDate,omitempty"`
	Tags                []string       `json:"tags,omitempty"`
	Flag                bool           `json:"flag,omitempty"`
	Tlp                 int            `json:"tlp,omitempty"`
	TlpLabel            string         `json:"tlpLabel,omitempty"`
	Pap                 int            `json:"pap,omitempty"`
	PapLabel            string         `json:"papLabel,omitempty"`
	Status              string         `json:"status,omitempty"`
	Stage               string         `json:"stage,omitempty"`
	Summary             string         `json:"summary,omitempty"`
	ImpactStatus        string         `json:"impactStatus,omitempty"`
	Assignee            string         `json:"assignee,omitempty"`
	Access              Access         `json:"access,omitempty"`
	CustomFields        []CustomFields `json:"customFields,omitempty"`
	UserPermissions     []string       `json:"userPermissions,omitempty"`
	ExtraData           ExtraData      `json:"extraData,omitempty"`
	NewDate             int64          `json:"newDate,omitempty"`
	InProgressDate      int64          `json:"inProgressDate,omitempty"`
	ClosedDate          int64          `json:"closedDate,omitempty"`
	AlertDate           int64          `json:"alertDate,omitempty"`
	AlertNewDate        int64          `json:"alertNewDate,omitempty"`
	AlertInProgressDate int64          `json:"alertInProgressDate,omitempty"`
	AlertImportedDate   int64          `json:"alertImportedDate,omitempty"`
	TimeToDetect        int            `json:"timeToDetect,omitempty"`
	TimeToTriage        int            `json:"timeToTriage,omitempty"`
	TimeToQualify       int            `json:"timeToQualify,omitempty"`
	TimeToAcknowledge   int            `json:"timeToAcknowledge,omitempty"`
	TimeToResolve       int            `json:"timeToResolve,omitempty"`
	HandlingDuration    int            `json:"handlingDuration,omitempty"`
}
type Access struct {
	Kind string `json:"_kind,omitempty"`
}
type CustomFields struct {
	ID    string `json:"_id,omitempty"`
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
	Value any    `json:"value,omitempty"`
	Order int    `json:"order,omitempty"`
}
type ExtraData interface {
}

//
