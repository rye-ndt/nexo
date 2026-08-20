package role_manager

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
	store   input_itf.RoleStorage
	archive input_itf.RoleArchive
}

func InitV1(store input_itf.RoleStorage, archive input_itf.RoleArchive) (core_itf.RoleManager, error) {
	if store == nil {
		return nil, custom_error.Critical("role storage is not initialized")
	}

	if archive == nil {
		return nil, custom_error.Critical("role archive is not initialized")
	}

	return &v1{
		store:   store,
		archive: archive,
	}, nil
}

func (m *v1) Upsert(role *core_itf.Role) (uuid.UUID, error) {
	if role == nil {
		return uuid.Nil, custom_error.Critical("role is empty")
	}

	id := role.ID

	if id == uuid.Nil {
		uid, err := uuid.NewV7()
		if err != nil {
			return uuid.Nil, custom_error.Critical("cannot create uuid: %v", err)
		}

		id = uid
	}

	now := helpers.NewUTC()

	entity := &input_itf.RoleEntity{
		ID:              id,
		Name:            strings.TrimSpace(role.Name),
		Description:     strings.TrimSpace(role.Description),
		Effort:          role.Effort,
		Retryable:       role.Retryable,
		PauseForReview:  role.PauseForReview,
		Inputs:          inputEntities(role.Inputs),
		Instructions:    prompts(role.Instructions),
		OutputStructure: strings.TrimSpace(role.OutputStructure),
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := helpers.ValidateStruct(entity); err != nil {
		return uuid.Nil, custom_error.Critical("invalid role %q: %v", entity.Name, err)
	}

	if err := m.store.Upsert(entity); err != nil {
		return uuid.Nil, custom_error.Critical("cannot save role %q: %v", entity.Name, err)
	}

	return id, nil
}

func (m *v1) List() ([]*core_itf.Role, error) {
	stored, err := m.store.List()
	if err != nil {
		return nil, custom_error.Critical("cannot list roles: %v", err)
	}

	roles := make([]*core_itf.Role, 0, len(stored))

	for _, entity := range stored {
		roles = append(roles, role(entity))
	}

	return roles, nil
}

func (m *v1) Get(id uuid.UUID) (*core_itf.Role, error) {
	stored, err := m.find(id)
	if err != nil {
		return nil, err
	}

	return role(stored), nil
}

func (m *v1) Remove(id uuid.UUID) error {
	if _, err := m.find(id); err != nil {
		return err
	}

	if err := m.store.Remove(id); err != nil {
		return custom_error.Critical("cannot remove role %v: %v", id, err)
	}

	return nil
}

func (m *v1) Export(ids []uuid.UUID, path string) (int, error) {
	if len(ids) == 0 {
		return 0, custom_error.Critical("pick at least one role to export")
	}

	records := make([]*input_itf.RoleRecord, 0, len(ids))
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

	doc := &input_itf.RoleExport{
		Version:    input_itf.ArchiveVersion,
		ExportedAt: helpers.NewUTC(),
		Roles:      records,
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
		return 0, custom_error.Critical("cannot list roles: %v", err)
	}

	if err := checkFile(doc.Roles); err != nil {
		return 0, err
	}

	if err := checkConflicts(doc.Roles, stored); err != nil {
		return 0, err
	}

	now := helpers.NewUTC()
	entities := make([]*input_itf.RoleEntity, 0, len(doc.Roles))

	for _, imported := range doc.Roles {
		entity := importedEntity(imported, now)

		if err := helpers.ValidateStruct(entity); err != nil {
			return 0, custom_error.TypedCritical(enums.ErrRoleFileInvalid,
				"invalid role %q: %v", entity.Name, err)
		}

		entities = append(entities, entity)
	}

	if err := m.store.UpsertMany(entities); err != nil {
		return 0, custom_error.Critical("cannot save the imported roles: %v", err)
	}

	return len(entities), nil
}

func checkFile(records []*input_itf.RoleRecord) error {
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

	return custom_error.TypedCritical(enums.ErrRoleFileInvalid, "%s", strings.Join(problems, "; "))
}

func checkConflicts(records []*input_itf.RoleRecord, stored []*input_itf.RoleEntity) error {
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
			problems = append(problems, fmt.Sprintf("a role named %q is already here", name))
			continue
		}

		if ids[imported.ID] {
			problems = append(problems, fmt.Sprintf("%q carries the id of a role already here", name))
		}
	}

	if len(problems) == 0 {
		return nil
	}

	return custom_error.TypedCritical(enums.ErrRoleConflict,
		"nothing was imported: %s", strings.Join(problems, "; "))
}

func (m *v1) find(id uuid.UUID) (*input_itf.RoleEntity, error) {
	if id == uuid.Nil {
		return nil, custom_error.Critical("role id is empty")
	}

	stored, err := m.store.Find(id)
	if err != nil {
		return nil, custom_error.Critical("cannot get role %v: %v", id, err)
	}

	if stored == nil {
		return nil, custom_error.Critical("role %v not found", id)
	}

	return stored, nil
}

func inputEntities(inputs map[string]*core_itf.RoleInputs) map[string]*input_itf.RoleInputEntity {
	entities := map[string]*input_itf.RoleInputEntity{}

	for key, input := range inputs {
		if input == nil {
			entities[strings.TrimSpace(key)] = nil
			continue
		}

		entities[strings.TrimSpace(key)] = &input_itf.RoleInputEntity{
			Description: input.Description,
			Required:    input.Required,
			Type:        input.Type,
			Default:     input.Default,
			Options:     input.Options,
		}
	}

	return entities
}

func prompts(instructions map[string]string) map[string]string {
	cloned := map[string]string{}

	for key, prompt := range instructions {
		cloned[strings.TrimSpace(key)] = strings.TrimSpace(prompt)
	}

	return cloned
}

func record(stored *input_itf.RoleEntity) *input_itf.RoleRecord {
	inputs := map[string]*input_itf.RoleInputRecord{}

	for key, input := range stored.Inputs {
		if input == nil {
			continue
		}

		inputs[key] = &input_itf.RoleInputRecord{
			Description: input.Description,
			Required:    input.Required,
			Type:        input.Type,
			Default:     input.Default,
			Options:     input.Options,
		}
	}

	return &input_itf.RoleRecord{
		ID:              stored.ID,
		Name:            stored.Name,
		Description:     stored.Description,
		Effort:          stored.Effort,
		Retryable:       stored.Retryable,
		PauseForReview:  stored.PauseForReview,
		Inputs:          inputs,
		Instructions:    stored.Instructions,
		OutputStructure: stored.OutputStructure,
	}
}

func importedEntity(imported *input_itf.RoleRecord, now time.Time) *input_itf.RoleEntity {
	inputs := map[string]*input_itf.RoleInputEntity{}

	for key, input := range imported.Inputs {
		if input == nil {
			continue
		}

		inputs[strings.TrimSpace(key)] = &input_itf.RoleInputEntity{
			Description: input.Description,
			Required:    input.Required,
			Type:        input.Type,
			Default:     input.Default,
			Options:     input.Options,
		}
	}

	return &input_itf.RoleEntity{
		ID:              imported.ID,
		Name:            strings.TrimSpace(imported.Name),
		Description:     strings.TrimSpace(imported.Description),
		Effort:          imported.Effort,
		Retryable:       imported.Retryable,
		PauseForReview:  imported.PauseForReview,
		Inputs:          inputs,
		Instructions:    prompts(imported.Instructions),
		OutputStructure: strings.TrimSpace(imported.OutputStructure),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

func role(entity *input_itf.RoleEntity) *core_itf.Role {
	inputs := map[string]*core_itf.RoleInputs{}

	for key, input := range entity.Inputs {
		inputs[key] = &core_itf.RoleInputs{
			Description: input.Description,
			Required:    input.Required,
			Type:        input.Type,
			Default:     input.Default,
			Options:     input.Options,
		}
	}

	return &core_itf.Role{
		ID:              entity.ID,
		Name:            entity.Name,
		Description:     entity.Description,
		Effort:          entity.Effort,
		Retryable:       entity.Retryable,
		PauseForReview:  entity.PauseForReview,
		Inputs:          inputs,
		Instructions:    prompts(entity.Instructions),
		OutputStructure: entity.OutputStructure,
	}
}
