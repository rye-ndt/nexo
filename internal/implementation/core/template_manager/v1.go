package template_manager

import (
	"fmt"
	"strings"
	"time"

	"hexago/internal/helpers"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"

	"github.com/google/uuid"
)

type v1 struct {
	store   input_itf.TemplateStorage
	archive input_itf.TemplateArchive
}

func InitV1(store input_itf.TemplateStorage, archive input_itf.TemplateArchive) (core_itf.AgentTemplateManager, error) {
	if store == nil {
		return nil, custom_error.Critical("template storage is not initialized")
	}

	if archive == nil {
		return nil, custom_error.Critical("template archive is not initialized")
	}

	return &v1{
		store:   store,
		archive: archive,
	}, nil
}

func (m *v1) Upsert(template *core_itf.Template) (uuid.UUID, error) {
	if template == nil {
		return uuid.Nil, custom_error.Critical("template is empty")
	}

	id := template.ID

	if id == uuid.Nil {
		uid, err := uuid.NewV7()
		if err != nil {
			return uuid.Nil, custom_error.Critical("cannot create uuid: %v", err)
		}

		id = uid
	}

	now := helpers.NewUTC()

	entity := &input_itf.TemplateEntity{
		ID:                   id,
		Name:                 strings.TrimSpace(template.Name),
		Role:                 strings.TrimSpace(template.Role),
		TaskLevel:            template.TaskLevel,
		Retryable:            template.Retryable,
		ManualAcceptRequired: template.ManualAcceptRequired,
		Params:               paramEntities(template.Params),
		SystemPrompts:        prompts(template.SystemPrompts),
		OutputStructure:      strings.TrimSpace(template.OutputStructure),
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	if err := helpers.ValidateStruct(entity); err != nil {
		return uuid.Nil, custom_error.Critical("invalid template %q: %v", entity.Name, err)
	}

	if err := m.store.Upsert(entity); err != nil {
		return uuid.Nil, custom_error.Critical("cannot save template %q: %v", entity.Name, err)
	}

	return id, nil
}

func (m *v1) List() ([]*core_itf.Template, error) {
	stored, err := m.store.List()
	if err != nil {
		return nil, custom_error.Critical("cannot list templates: %v", err)
	}

	templates := make([]*core_itf.Template, 0, len(stored))

	for _, entity := range stored {
		templates = append(templates, template(entity))
	}

	return templates, nil
}

func (m *v1) Get(id uuid.UUID) (*core_itf.Template, error) {
	stored, err := m.find(id)
	if err != nil {
		return nil, err
	}

	return template(stored), nil
}

func (m *v1) Remove(id uuid.UUID) error {
	if _, err := m.find(id); err != nil {
		return err
	}

	if err := m.store.Remove(id); err != nil {
		return custom_error.Critical("cannot remove template %v: %v", id, err)
	}

	return nil
}

func (m *v1) Export(ids []uuid.UUID, path string) (int, error) {
	if len(ids) == 0 {
		return 0, custom_error.Critical("pick at least one template to export")
	}

	records := make([]*input_itf.TemplateRecord, 0, len(ids))
	taken := map[uuid.UUID]bool{}

	for _, id := range ids {
		if taken[id] {
			continue
		}

		stored, err := m.find(id)
		if err != nil {
			return 0, err
		}

		taken[id] = true

		records = append(records, record(stored))
	}

	doc := &input_itf.TemplateExport{
		Version:    input_itf.ArchiveVersion,
		ExportedAt: helpers.NewUTC(),
		Templates:  records,
	}

	if err := m.archive.Write(path, doc); err != nil {
		return 0, err
	}

	return len(records), nil
}

func (m *v1) Import(path string) (int, error) {
	doc, err := m.archive.Read(path)
	if err != nil {
		return 0, err
	}

	stored, err := m.store.List()
	if err != nil {
		return 0, custom_error.Critical("cannot list templates: %v", err)
	}

	if err := checkFile(doc.Templates); err != nil {
		return 0, err
	}

	if err := checkConflicts(doc.Templates, stored); err != nil {
		return 0, err
	}

	now := helpers.NewUTC()
	entities := make([]*input_itf.TemplateEntity, 0, len(doc.Templates))

	for _, imported := range doc.Templates {
		entity := importedEntity(imported, now)

		if err := helpers.ValidateStruct(entity); err != nil {
			return 0, custom_error.TypedCritical(enums.ErrTemplateFileInvalid,
				"invalid template %q: %v", entity.Name, err)
		}

		entities = append(entities, entity)
	}

	if err := m.store.UpsertMany(entities); err != nil {
		return 0, custom_error.Critical("cannot save the imported templates: %v", err)
	}

	return len(entities), nil
}

func checkFile(records []*input_itf.TemplateRecord) error {
	problems := []string{}
	names := map[string]bool{}
	ids := map[uuid.UUID]bool{}

	for _, imported := range records {
		name := strings.TrimSpace(imported.Name)

		if names[strings.ToLower(name)] {
			problems = append(problems, fmt.Sprintf("the file lists %q twice", name))
		}

		if ids[imported.ID] {
			problems = append(problems, fmt.Sprintf("the file lists the id %s twice", imported.ID))
		}

		names[strings.ToLower(name)] = true
		ids[imported.ID] = true
	}

	if len(problems) == 0 {
		return nil
	}

	return custom_error.TypedCritical(enums.ErrTemplateFileInvalid, "%s", strings.Join(problems, "; "))
}

func checkConflicts(records []*input_itf.TemplateRecord, stored []*input_itf.TemplateEntity) error {
	problems := []string{}
	names := map[string]bool{}
	ids := map[uuid.UUID]bool{}

	for _, existing := range stored {
		names[strings.ToLower(strings.TrimSpace(existing.Name))] = true
		ids[existing.ID] = true
	}

	for _, imported := range records {
		name := strings.TrimSpace(imported.Name)

		if names[strings.ToLower(name)] {
			problems = append(problems, fmt.Sprintf("a template named %q is already here", name))
			continue
		}

		if ids[imported.ID] {
			problems = append(problems, fmt.Sprintf("%q carries the id of a template already here", name))
		}
	}

	if len(problems) == 0 {
		return nil
	}

	return custom_error.TypedCritical(enums.ErrTemplateConflict,
		"nothing was imported: %s", strings.Join(problems, "; "))
}

func (m *v1) find(id uuid.UUID) (*input_itf.TemplateEntity, error) {
	if id == uuid.Nil {
		return nil, custom_error.Critical("template id is empty")
	}

	stored, err := m.store.Find(id)
	if err != nil {
		return nil, custom_error.Critical("cannot get template %v: %v", id, err)
	}

	if stored == nil {
		return nil, custom_error.Critical("template %v not found", id)
	}

	return stored, nil
}

func paramEntities(params map[string]*core_itf.TemplateParams) map[string]*input_itf.TemplateParamEntity {
	entities := map[string]*input_itf.TemplateParamEntity{}

	for key, param := range params {
		if param == nil {
			entities[strings.TrimSpace(key)] = nil
			continue
		}

		entities[strings.TrimSpace(key)] = &input_itf.TemplateParamEntity{
			Description: param.Description,
			Required:    param.Required,
			Type:        param.Type,
			Default:     param.Default,
			Options:     param.Options,
		}
	}

	return entities
}

func prompts(systemPrompts map[string]string) map[string]string {
	cloned := map[string]string{}

	for key, prompt := range systemPrompts {
		cloned[strings.TrimSpace(key)] = strings.TrimSpace(prompt)
	}

	return cloned
}

func record(stored *input_itf.TemplateEntity) *input_itf.TemplateRecord {
	params := map[string]*input_itf.TemplateParamRecord{}

	for key, param := range stored.Params {
		if param == nil {
			continue
		}

		params[key] = &input_itf.TemplateParamRecord{
			Description: param.Description,
			Required:    param.Required,
			Type:        param.Type,
			Default:     param.Default,
			Options:     param.Options,
		}
	}

	return &input_itf.TemplateRecord{
		ID:                   stored.ID,
		Name:                 stored.Name,
		Role:                 stored.Role,
		TaskLevel:            stored.TaskLevel,
		Retryable:            stored.Retryable,
		ManualAcceptRequired: stored.ManualAcceptRequired,
		Params:               params,
		SystemPrompts:        stored.SystemPrompts,
		OutputStructure:      stored.OutputStructure,
	}
}

func importedEntity(imported *input_itf.TemplateRecord, now time.Time) *input_itf.TemplateEntity {
	params := map[string]*input_itf.TemplateParamEntity{}

	for key, param := range imported.Params {
		if param == nil {
			continue
		}

		params[strings.TrimSpace(key)] = &input_itf.TemplateParamEntity{
			Description: param.Description,
			Required:    param.Required,
			Type:        param.Type,
			Default:     param.Default,
			Options:     param.Options,
		}
	}

	return &input_itf.TemplateEntity{
		ID:                   imported.ID,
		Name:                 strings.TrimSpace(imported.Name),
		Role:                 strings.TrimSpace(imported.Role),
		TaskLevel:            imported.TaskLevel,
		Retryable:            imported.Retryable,
		ManualAcceptRequired: imported.ManualAcceptRequired,
		Params:               params,
		SystemPrompts:        prompts(imported.SystemPrompts),
		OutputStructure:      strings.TrimSpace(imported.OutputStructure),
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

func template(entity *input_itf.TemplateEntity) *core_itf.Template {
	params := map[string]*core_itf.TemplateParams{}

	for key, param := range entity.Params {
		params[key] = &core_itf.TemplateParams{
			Description: param.Description,
			Required:    param.Required,
			Type:        param.Type,
			Default:     param.Default,
			Options:     param.Options,
		}
	}

	return &core_itf.Template{
		ID:                   entity.ID,
		Name:                 entity.Name,
		Role:                 entity.Role,
		TaskLevel:            entity.TaskLevel,
		Retryable:            entity.Retryable,
		ManualAcceptRequired: entity.ManualAcceptRequired,
		Params:               params,
		SystemPrompts:        prompts(entity.SystemPrompts),
		OutputStructure:      entity.OutputStructure,
	}
}
