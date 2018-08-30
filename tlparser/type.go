package tlparser

type Schema struct {
    Types     []*Type     `json:"types"`
    Classes   []*Class    `json:"classes"`
    Functions []*Function `json:"functions"`
}

type Type struct {
    Name        string      `json:"name"`
    Description string      `json:"description"`
    Class       string      `json:"class"`
    Properties  []*Property `json:"properties"`
}

type Class struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

type Function struct {
    Name          string      `json:"name"`
    Description   string      `json:"description"`
    Class         string      `json:"class"`
    Properties    []*Property `json:"properties"`
    IsSynchronous bool        `json:"is_synchronous"`
}

type Property struct {
    Name        string `json:"name"`
    Type        string `json:"type"`
    Description string `json:"description"`
}
