package platform

type Variable struct {
	ID   VariableID
	Name string
}

type VariableID struct {
	Workspace   string
	Project     string
	Environment string
	Secret      string
	Name        string
}

func (v VariableID) EnvironmentID() EnvironmentID {
	return EnvironmentID{Workspace: v.Workspace, Project: v.Project, Name: v.Environment}
}
