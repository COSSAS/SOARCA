package schemas

type Task struct {
	Title       string `bson:"title" json:"title" validate:"required" example:"Task 1"`
	Group       string `bson:"group,omitempty" json:"group,omitempty" example:"Group 1"`
	Description string `bson:"description,omitempty" json:"description,omitempty" example:"Description of task 1"`
	Status      string `bson:"status,omitempty" json:"status,omitempty" example:"Open"`
	Flag        bool   `bson:"flag,omitempty" json:"flag,omitempty" example:"true"`
	StartDate   int64  `bson:"startDate,omitempty" json:"startDate,omitempty" example:"1640000000000"`
	EndDate     int64  `bson:"endDate,omitempty" json:"endDate,omitempty" example:"1640000000000"`
	Order       int    `bson:"order,omitempty" json:"order,omitempty" example:"1"`
	DueDate     int64  `bson:"dueDate,omitempty" json:"dueDate,omitempty" example:"1640000000000"`
	Assignee    string `bson:"assignee,omitempty" json:"assignee,omitempty" example:"Jane Doe"`
	Mandatory   bool   `bson:"mandatory,omitempty" json:"mandatory,omitempty" example:"true"`
}

type Page struct {
	Title    string `bson:"title" json:"title" example:"Page 1"`
	Content  string `bson:"content" json:"content" example:"Content of page 1"`
	Order    int    `bson:"order" json:"order" example:"1"`
	Category string `bson:"category" json:"category" example:"Category 1"`
}

type SharingParameter struct {
	Organisation   string `bson:"organisation" json:"organisation" example:"~354"`
	Share          bool   `bson:"share" json:"share" example:"true"`
	Profile        string `bson:"profile" json:"profile" example:"analyst"`
	TaskRule       string `bson:"taskRule" json:"taskRule" example:"Sharing rule applied on the case"`
	ObservableRule string `bson:"observableRule" json:"observableRule" example:"Sharing rule applied on the case"`
}

type Case struct {
	Title             string             `bson:"title" json:"title" validate:"required,min=1,max=512" example:"Example Case"`
	Description       string             `bson:"description" json:"description" validate:"required,max=1048576"`
	Severity          int                `bson:"severity,omitempty" json:"severity,omitempty" validate:"min=1,max=4" example:"2"`
	StartDate         int64              `bson:"startDate,omitempty" json:"startDate,omitempty" example:"1640000000000"`
	EndDate           int64              `bson:"endDate,omitempty" json:"endDate,omitempty" example:"1640000000000"`
	Tags              []string           `bson:"tags,omitempty" json:"tags,omitempty" validate:"max=128,dive,min=1,max=128" example:"[\"example\", \"test\"]"`
	Flag              bool               `bson:"flag,omitempty" json:"flag,omitempty" example:"false"`
	TLP               int                `bson:"tlp,omitempty" json:"tlp,omitempty" validate:"min=0,max=4" example:"2"`
	PAP               int                `bson:"pap,omitempty" json:"pap,omitempty" validate:"min=0,max=3" example:"2"`
	Status            string             `bson:"status,omitempty" json:"status,omitempty" validate:"min=1,max=64" example:"New"`
	Summary           string             `bson:"summary,omitempty" json:"summary,omitempty" validate:"max=1048576" example:"Summary of the case"`
	Assignee          string             `bson:"assignee,omitempty" json:"assignee,omitempty" validate:"max=128" example:"John Doe"`
	CustomFields      interface{}        `bson:"customFields,omitempty" json:"customFields,omitempty" example:"{\"property1\":null,\"property2\":null}"`
	CaseTemplate      string             `bson:"caseTemplate,omitempty" json:"caseTemplate,omitempty" validate:"max=128" example:"Template1"`
	Tasks             []Task             `bson:"tasks,omitempty" json:"tasks,omitempty"`
	Pages             []Page             `bson:"pages,omitempty" json:"pages,omitempty"`
	SharingParameters []SharingParameter `bson:"sharingParameters,omitempty" json:"sharingParameters,omitempty"`
	TaskRule          string             `bson:"taskRule,omitempty" json:"taskRule,omitempty" validate:"max=128" example:"Task rule"`
	ObservableRule    string             `bson:"observableRule,omitempty" json:"observableRule,omitempty" validate:"max=128" example:"Observable rule"`
}
