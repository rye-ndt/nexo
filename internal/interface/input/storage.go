package input_itf

type Storage interface {
	HarnessStorage
	TaskStore() TaskStorage
	MCPStore() StorageMCP
	TemplateStore() TemplateStorage
	DraftStore() DraftStorage
}
