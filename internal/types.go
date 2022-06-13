package internal

type WorkRequest struct {
	RequestType   RequestType `json:"requestType"`
	Kit           Kit         `json:"kit"`
	RepositoryUrl string      `json:"repositoryUrl"`
	Ref           string      `json:"ref"`
	Variables     []Variable  `json:"variables"`
	Hash          string      `json:"hash"`
}

type Kit struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type RequestType string

const (
	PreCompile RequestType = "precompile"
	Compile    RequestType = "compile"
	Deploy     RequestType = "deploy"
)

type Buildkit struct {
	Url       string
	Dir       string
	Variables []Variable
}

type DeployKit struct {
	Url       string
	Dir       string
	Variables []Variable
}

type VariableType string

const (
	StringVariable VariableType = "string"
	SecretVariable VariableType = "secret"
)

type Variable struct {
	Key   string       `json:"key"`
	Value string       `json:"value"`
	Type  VariableType `json:"type"`
}
