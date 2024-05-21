package api

type Error struct {
	Status         int    `json:"status" validate:"required" example:"400"`
	Message        string `json:"message" validate:"required" example:"missing argument in call"`
	OriginalCall   string `json:"original-call" validate:"required" example:"/example/route"`
	DownstreamCall string `json:"downstream-call" validate:"omitempty" example:"{"some" : "json"}"`
}
