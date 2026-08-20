package input_itf

type Storage interface {
	HarnessStorage
	StepStore() StepStorage
	MCPStore() StorageMCP
	RoleStore() RoleStorage
	DraftStore() DraftStorage
}
