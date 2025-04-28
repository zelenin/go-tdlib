package tlparser

type Schema struct {
	Constructors []*Constructor `json:"constructors"`
	Types        []*Type        `json:"types"`
	Functions    []*Function    `json:"functions"`
}

type Constructor struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Args        []*Arg `json:"args"`
	ResultType  string `json:"result_type"`
}

type Type struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type FunctionType string

const (
	FUNCTION_TYPE_COMMON FunctionType = "common"
	FUNCTION_TYPE_USER   FunctionType = "user"
	FUNCTION_TYPE_BOT    FunctionType = "bot"
)

type Function struct {
	Name          string       `json:"name"`
	Description   string       `json:"description"`
	Args          []*Arg       `json:"args"`
	ResultType    string       `json:"result_type"`
	IsSynchronous bool         `json:"is_synchronous"`
	Type          FunctionType `json:"type"`
}

type Arg struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
}
