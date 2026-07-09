package expression

type Query string

type IExpression interface {
	Execute(string, Query) (string, error)
	GetEngineName() string
}
