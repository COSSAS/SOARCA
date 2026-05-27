package expression

type Query string

type IExpression interface {
	Execute(string, Query) (error, string)
	GetEngineName() string
}
