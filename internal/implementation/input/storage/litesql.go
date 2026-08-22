package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	input_itf "hexago/internal/interface/input"
)

// driverName is modernc's, and is deliberately not "sqlite3" — that is the cgo one.
const driverName = "sqlite"

var migrations = []string{
	`CREATE TABLE harnesses (
		name TEXT PRIMARY KEY,
		version TEXT NOT NULL,
		platform TEXT NOT NULL,
		path TEXT NOT NULL
	)`,
	`CREATE TABLE workflows (
		id TEXT PRIMARY KEY,
		started_at TEXT NOT NULL,
		completed_at TEXT NOT NULL,
		total_step INTEGER NOT NULL,
		total_retry INTEGER NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		project_dir_path TEXT NOT NULL DEFAULT ''
	)`,
	`CREATE TABLE steps (
		id TEXT PRIMARY KEY,
		workflow_id TEXT NOT NULL,
		name TEXT NOT NULL,
		preferred_model TEXT NOT NULL,
		extra_guidance TEXT NOT NULL,
		retry_count INTEGER NOT NULL,
		status TEXT NOT NULL,
		last_report_id TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		thinking_level TEXT NOT NULL DEFAULT '',
		instructions TEXT NOT NULL DEFAULT '[]',
		auto_retry INTEGER NOT NULL DEFAULT 0,
		depends_on_step_ids TEXT NOT NULL DEFAULT '[]',
		pause_for_review INTEGER NOT NULL DEFAULT 0,
		effort TEXT NOT NULL DEFAULT ''
	)`,
	`CREATE TABLE step_results (
		id TEXT PRIMARY KEY,
		step_id TEXT NOT NULL,
		agent_id TEXT NOT NULL,
		attempt_status TEXT NOT NULL,
		handoff TEXT NOT NULL,
		started_at TEXT NOT NULL,
		completed_at TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		context_usage TEXT NOT NULL DEFAULT ''
	)`,
	`CREATE TABLE mcp_credentials (
		name TEXT PRIMARY KEY,
		client_id TEXT NOT NULL,
		token_endpoint TEXT NOT NULL,
		encrypted_oauth_key TEXT NOT NULL,
		expired_at TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		account TEXT NOT NULL DEFAULT ''
	)`,
	`CREATE TABLE roles (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		description TEXT NOT NULL,
		effort TEXT NOT NULL,
		retryable INTEGER NOT NULL,
		inputs TEXT NOT NULL,
		instructions TEXT NOT NULL,
		created_at TEXT NOT NULL,
		updated_at TEXT NOT NULL,
		pause_for_review INTEGER NOT NULL DEFAULT 0,
		output_structure TEXT NOT NULL DEFAULT ''
	)`,
	`CREATE TABLE workflow_drafts (
		id TEXT PRIMARY KEY,
		doc TEXT NOT NULL,
		updated_at TEXT NOT NULL
	)`,
	seedRoles,
	seedWorkflow,
}

const roleColumns = `id, name, description, effort, retryable, pause_for_review, inputs, instructions, output_structure, created_at, updated_at`

type litesql struct {
	db *sql.DB
}

type stepStore struct {
	db *sql.DB
}

type mcpStore struct {
	db *sql.DB
}

type roleStore struct {
	db *sql.DB
}

type draftStore struct {
	db *sql.DB
}

type execer interface {
	Exec(query string, args ...any) (sql.Result, error)
}

func New(path string) (input_itf.Storage, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open(driverName, dsn(path))
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}

	return &litesql{db: db}, nil
}

// dsn asks for WAL journalling and a busy timeout. database/sql hands out several
// connections, and step writes now come from the coordinator, the heartbeat
// watcher and every MCP report at once; without these two pragmas the second
// writer of any overlapping pair fails instantly with SQLITE_BUSY instead of
// waiting its turn. The path is escaped because the macOS data dir has a space in it.
func dsn(path string) string {
	// SQLite reads everything between file:// and the first slash as an
	// authority, so a Windows path has to arrive as /C:/Users/... — the raw
	// C:\Users\... becomes file://C:%5CUsers%5C... and is rejected as one.
	slashed := filepath.ToSlash(path)
	if !strings.HasPrefix(slashed, "/") {
		slashed = "/" + slashed
	}

	uri := url.URL{Scheme: "file", Path: slashed}
	uri.RawQuery = url.Values{
		"_pragma": {"busy_timeout(5000)", "journal_mode(WAL)"},
	}.Encode()

	return uri.String()
}

func migrate(db *sql.DB) error {
	version, err := alignedVersion(db)
	if err != nil {
		return err
	}

	for i := version; i < len(migrations); i++ {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(migrations[i]); err != nil {
			tx.Rollback()
			return custom_error.Critical("migration %d: %v", i+1, err)
		}
		if _, err := tx.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, i+1)); err != nil {
			tx.Rollback()
			return custom_error.Critical("migration %d: %v", i+1, err)
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

var collapsedTables = []string{
	"harnesses",
	"workflows",
	"steps",
	"step_results",
	"mcp_credentials",
	"roles",
	"workflow_drafts",
}

func alignedVersion(db *sql.DB) (int, error) {
	var version int
	if err := db.QueryRow(`PRAGMA user_version`).Scan(&version); err != nil {
		return 0, err
	}

	if version <= len(migrations) {
		return version, nil
	}

	built, err := builtByTheCollapsedList(db)
	if err != nil {
		return 0, err
	}

	if !built {
		return 0, custom_error.Critical(
			"database is at version %d and predates the collapsed schema, which cannot upgrade it: "+
				"move or delete the app data directory and start again",
			version,
		)
	}

	if _, err := db.Exec(fmt.Sprintf(`PRAGMA user_version = %d`, len(migrations))); err != nil {
		return 0, err
	}

	return len(migrations), nil
}

func builtByTheCollapsedList(db *sql.DB) (bool, error) {
	for _, table := range collapsedTables {
		var name string

		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`, table).Scan(&name)

		switch {
		case errors.Is(err, sql.ErrNoRows):
			return false, nil
		case err != nil:
			return false, err
		}
	}

	return true, nil
}

func (s *litesql) Save(info *input_itf.HarnessEntity) error {
	_, err := s.db.Exec(`INSERT INTO harnesses (name, version, platform, path)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			version = excluded.version,
			platform = excluded.platform,
			path = excluded.path`,
		info.Name,
		info.Version,
		info.Platform.String(),
		info.Path,
	)
	return err
}

func (s *litesql) Find(name string) (*input_itf.HarnessEntity, error) {
	info := &input_itf.HarnessEntity{}
	var platform string

	err := s.db.QueryRow(`SELECT name, version, platform, path
		FROM harnesses WHERE name = ?`, name).
		Scan(&info.Name, &info.Version, &platform, &info.Path)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	info.Platform = enums.OS(platform)
	return info, nil
}

func (s *litesql) StepStore() input_itf.StepStorage {
	return &stepStore{db: s.db}
}

func (s *litesql) MCPStore() input_itf.StorageMCP {
	return &mcpStore{db: s.db}
}

func (s *litesql) RoleStore() input_itf.RoleStorage {
	return &roleStore{db: s.db}
}

func (s *litesql) DraftStore() input_itf.DraftStorage {
	return &draftStore{db: s.db}
}

func (s *draftStore) List() ([]*input_itf.WorkflowDraftEntity, error) {
	rows, err := s.db.Query(`SELECT id, doc, updated_at FROM workflow_drafts ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*input_itf.WorkflowDraftEntity{}

	for rows.Next() {
		d := &input_itf.WorkflowDraftEntity{}
		var id, updatedAt string

		if err := rows.Scan(&id, &d.Doc, &updatedAt); err != nil {
			return nil, err
		}

		d.ID = parseUUID(id)
		d.UpdatedAt = parseTime(updatedAt)

		list = append(list, d)
	}

	return list, rows.Err()
}

func (s *draftStore) Save(draft *input_itf.WorkflowDraftEntity) error {
	_, err := s.db.Exec(`INSERT INTO workflow_drafts (id, doc, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			doc = excluded.doc,
			updated_at = excluded.updated_at`,
		draft.ID.String(),
		draft.Doc,
		formatTime(draft.UpdatedAt),
	)
	return err
}

func (s *draftStore) Delete(id uuid.UUID) error {
	_, err := s.db.Exec(`DELETE FROM workflow_drafts WHERE id = ?`, id.String())
	return err
}

func (s *roleStore) Upsert(t *input_itf.RoleEntity) error {
	return upsertRole(s.db, t)
}

func (s *roleStore) UpsertMany(roles []*input_itf.RoleEntity) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, t := range roles {
		if err := upsertRole(tx, t); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func upsertRole(db execer, t *input_itf.RoleEntity) error {
	inputs, err := json.Marshal(t.Inputs)
	if err != nil {
		return err
	}

	instructions, err := json.Marshal(t.Instructions)
	if err != nil {
		return err
	}

	_, err = db.Exec(`INSERT INTO roles (`+roleColumns+`)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			effort = excluded.effort,
			retryable = excluded.retryable,
			pause_for_review = excluded.pause_for_review,
			inputs = excluded.inputs,
			instructions = excluded.instructions,
			output_structure = excluded.output_structure,
			updated_at = excluded.updated_at`,
		t.ID.String(),
		t.Name,
		t.Description,
		t.Effort.String(),
		t.Retryable,
		t.PauseForReview,
		string(inputs),
		string(instructions),
		t.OutputStructure,
		formatTime(t.CreatedAt),
		formatTime(t.UpdatedAt),
	)
	return err
}

func (s *roleStore) List() ([]*input_itf.RoleEntity, error) {
	rows, err := s.db.Query(`SELECT ` + roleColumns + ` FROM roles ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*input_itf.RoleEntity{}

	for rows.Next() {
		t, err := scanRole(rows.Scan)
		if err != nil {
			return nil, err
		}

		list = append(list, t)
	}

	return list, rows.Err()
}

func (s *roleStore) Find(id uuid.UUID) (*input_itf.RoleEntity, error) {
	row := s.db.QueryRow(`SELECT `+roleColumns+` FROM roles WHERE id = ?`, id.String())

	t, err := scanRole(row.Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *roleStore) Remove(id uuid.UUID) error {
	_, err := s.db.Exec(`DELETE FROM roles WHERE id = ?`, id.String())
	return err
}

func scanRole(scan func(dest ...any) error) (*input_itf.RoleEntity, error) {
	t := &input_itf.RoleEntity{}

	var id, effort, inputs, instructions, createdAt, updatedAt string

	if err := scan(&id, &t.Name, &t.Description, &effort, &t.Retryable, &t.PauseForReview,
		&inputs, &instructions, &t.OutputStructure, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	t.ID = parseUUID(id)
	t.Effort = enums.Effort(effort)
	t.CreatedAt = parseTime(createdAt)
	t.UpdatedAt = parseTime(updatedAt)

	if err := json.Unmarshal([]byte(inputs), &t.Inputs); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(instructions), &t.Instructions); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *mcpStore) ListAuthenticated() ([]*input_itf.MCPEntity, error) {
	rows, err := s.db.Query(`SELECT name, client_id, token_endpoint, encrypted_oauth_key,
		account, expired_at, created_at, updated_at
		FROM mcp_credentials WHERE encrypted_oauth_key != ''`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*input_itf.MCPEntity{}

	for rows.Next() {
		m := &input_itf.MCPEntity{}
		var expiredAt, createdAt, updatedAt string

		if err := rows.Scan(&m.Name, &m.ClientID, &m.TokenEndpoint, &m.EncryptedOAuthKey,
			&m.Account, &expiredAt, &createdAt, &updatedAt); err != nil {
			return nil, err
		}

		m.ExpiredAt = parseTime(expiredAt)
		m.CreatedAt = parseTime(createdAt)
		m.UpdatedAt = parseTime(updatedAt)

		list = append(list, m)
	}

	return list, rows.Err()
}

func (s *mcpStore) UpsertCredentials(mcp *input_itf.MCPEntity) error {
	_, err := s.db.Exec(`INSERT INTO mcp_credentials
		(name, client_id, token_endpoint, encrypted_oauth_key,
		account, expired_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			client_id = excluded.client_id,
			token_endpoint = excluded.token_endpoint,
			encrypted_oauth_key = excluded.encrypted_oauth_key,
			account = excluded.account,
			expired_at = excluded.expired_at,
			updated_at = excluded.updated_at`,
		mcp.Name,
		mcp.ClientID,
		mcp.TokenEndpoint,
		mcp.EncryptedOAuthKey,
		mcp.Account,
		formatTime(mcp.ExpiredAt),
		formatTime(mcp.CreatedAt),
		formatTime(mcp.UpdatedAt),
	)
	return err
}

func (s *mcpStore) GetCredentials(name string) (*input_itf.MCPEntity, error) {
	m := &input_itf.MCPEntity{}
	var expiredAt, createdAt, updatedAt string

	err := s.db.QueryRow(`SELECT name, client_id, token_endpoint, encrypted_oauth_key,
		account, expired_at, created_at, updated_at
		FROM mcp_credentials WHERE name = ?`, name).
		Scan(&m.Name, &m.ClientID, &m.TokenEndpoint, &m.EncryptedOAuthKey,
			&m.Account, &expiredAt, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	m.ExpiredAt = parseTime(expiredAt)
	m.CreatedAt = parseTime(createdAt)
	m.UpdatedAt = parseTime(updatedAt)

	return m, nil
}

func (s *mcpStore) DeleteCredentials(name string) error {
	_, err := s.db.Exec(`DELETE FROM mcp_credentials WHERE name = ?`, name)
	return err
}

func (s *stepStore) SaveStepHistory(
	workflows []*input_itf.WorkflowEntity,
	steps []*input_itf.StepEntity,
	reports []*input_itf.StepResultEntity,
) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, sess := range workflows {
		if err := saveWorkflow(tx, sess); err != nil {
			return err
		}
	}

	for _, t := range steps {
		if err := saveStep(tx, t); err != nil {
			return err
		}
	}

	for _, r := range reports {
		if err := saveReport(tx, r); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *stepStore) LoadStepHistory() ([]*input_itf.WorkflowSnapshot, error) {
	workflows, err := loadAll(s.db, `SELECT id, project_dir_path, started_at,
		completed_at, total_step, total_retry, created_at, updated_at
		FROM workflows ORDER BY created_at, id`, scanWorkflow)
	if err != nil {
		return nil, err
	}

	steps, err := loadAll(s.db, `SELECT id, workflow_id, name, effort, preferred_model,
		thinking_level, instructions, auto_retry, pause_for_review, extra_guidance,
		retry_count, status, depends_on_step_ids, last_report_id, created_at, updated_at
		FROM steps ORDER BY created_at, id`, scanStep)
	if err != nil {
		return nil, err
	}

	reports, err := loadAll(s.db, `SELECT id, step_id, agent_id, attempt_status, handoff,
		context_usage, started_at, completed_at, created_at, updated_at
		FROM step_results ORDER BY created_at, id`, scanReport)
	if err != nil {
		return nil, err
	}

	stepsOf := map[uuid.UUID][]*input_itf.StepEntity{}
	for _, t := range steps {
		stepsOf[t.WorkflowID] = append(stepsOf[t.WorkflowID], t)
	}

	reportsOf := map[uuid.UUID][]*input_itf.StepResultEntity{}
	for _, r := range reports {
		reportsOf[r.StepID] = append(reportsOf[r.StepID], r)
	}

	snapshots := make([]*input_itf.WorkflowSnapshot, 0, len(workflows))

	for _, sess := range workflows {
		snapshot := &input_itf.WorkflowSnapshot{Workflow: sess, Steps: stepsOf[sess.ID]}

		for _, t := range snapshot.Steps {
			snapshot.Reports = append(snapshot.Reports, reportsOf[t.ID]...)
		}

		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

func loadAll[T any](db *sql.DB, query string, scan func(*sql.Rows) (*T, error)) ([]*T, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []*T{}

	for rows.Next() {
		item, err := scan(rows)
		if err != nil {
			return nil, err
		}

		list = append(list, item)
	}

	return list, rows.Err()
}

func scanWorkflow(rows *sql.Rows) (*input_itf.WorkflowEntity, error) {
	sess := &input_itf.WorkflowEntity{}

	var id, startedAt, completedAt, createdAt, updatedAt string

	if err := rows.Scan(&id, &sess.ProjectDirPath, &startedAt, &completedAt,
		&sess.TotalStep, &sess.TotalRetry, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	sess.ID = parseUUID(id)
	sess.StartedAt = parseTime(startedAt)
	sess.CompletedAt = parseTime(completedAt)
	sess.CreatedAt = parseTime(createdAt)
	sess.UpdatedAt = parseTime(updatedAt)

	return sess, nil
}

func scanStep(rows *sql.Rows) (*input_itf.StepEntity, error) {
	t := &input_itf.StepEntity{}

	var id, workflowID, effort, preferredModel, thinkingLevel, instructions string
	var status, dependsOnStepIDs, lastReportID, createdAt, updatedAt string

	if err := rows.Scan(&id, &workflowID, &t.Name, &effort, &preferredModel, &thinkingLevel,
		&instructions, &t.AutoRetry, &t.PauseForReview, &t.ExtraGuidance,
		&t.RetryCount, &status, &dependsOnStepIDs, &lastReportID,
		&createdAt, &updatedAt); err != nil {
		return nil, err
	}

	t.ID = parseUUID(id)
	t.WorkflowID = parseUUID(workflowID)
	t.Effort = enums.Effort(effort)
	t.PreferredModel = enums.ModelName(preferredModel)
	t.ThinkingLevel = enums.ThinkingLevel(thinkingLevel)
	t.Status = enums.StepStatus(status)
	t.LastReportID = parseUUID(lastReportID)
	t.CreatedAt = parseTime(createdAt)
	t.UpdatedAt = parseTime(updatedAt)

	if err := json.Unmarshal([]byte(instructions), &t.Instructions); err != nil {
		return nil, err
	}

	return t, json.Unmarshal([]byte(dependsOnStepIDs), &t.DependsOnStepIDs)
}

func scanReport(rows *sql.Rows) (*input_itf.StepResultEntity, error) {
	r := &input_itf.StepResultEntity{}

	var id, stepID, agentID, attemptStatus, handoff, contextUsage string
	var startedAt, completedAt, createdAt, updatedAt string

	if err := rows.Scan(&id, &stepID, &agentID, &attemptStatus, &handoff, &contextUsage,
		&startedAt, &completedAt, &createdAt, &updatedAt); err != nil {
		return nil, err
	}

	r.ID = parseUUID(id)
	r.StepID = parseUUID(stepID)
	r.AgentID = parseUUID(agentID)
	r.AttemptStatus = enums.StepStatus(attemptStatus)
	r.StartedAt = parseTime(startedAt)
	r.CompletedAt = parseTime(completedAt)
	r.CreatedAt = parseTime(createdAt)
	r.UpdatedAt = parseTime(updatedAt)

	if err := json.Unmarshal([]byte(handoff), &r.Handoffs); err != nil {
		return nil, err
	}

	if contextUsage == "" {
		return r, nil
	}

	usage := &input_itf.ContextUsage{}
	if err := json.Unmarshal([]byte(contextUsage), usage); err != nil {
		return nil, err
	}

	r.ContextUsage = usage

	return r, nil
}

func saveWorkflow(e execer, sess *input_itf.WorkflowEntity) error {
	_, err := e.Exec(`INSERT OR REPLACE INTO workflows
		(id, project_dir_path, started_at, completed_at,
		total_step, total_retry, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		sess.ID.String(),
		sess.ProjectDirPath,
		formatTime(sess.StartedAt),
		formatTime(sess.CompletedAt),
		sess.TotalStep,
		sess.TotalRetry,
		formatTime(sess.CreatedAt),
		formatTime(sess.UpdatedAt),
	)
	return err
}

func saveStep(e execer, t *input_itf.StepEntity) error {
	instructions, err := json.Marshal(t.Instructions)
	if err != nil {
		return err
	}

	dependsOnIDs, err := json.Marshal(t.DependsOnStepIDs)
	if err != nil {
		return err
	}

	_, err = e.Exec(`INSERT OR REPLACE INTO steps
		(id, workflow_id, name, effort, preferred_model, thinking_level, instructions,
		auto_retry, pause_for_review,
		extra_guidance, retry_count, status, depends_on_step_ids,
		last_report_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		t.ID.String(),
		t.WorkflowID.String(),
		t.Name,
		string(t.Effort),
		string(t.PreferredModel),
		string(t.ThinkingLevel),
		string(instructions),
		t.AutoRetry,
		t.PauseForReview,
		t.ExtraGuidance,
		t.RetryCount,
		string(t.Status),
		string(dependsOnIDs),
		t.LastReportID.String(),
		formatTime(t.CreatedAt),
		formatTime(t.UpdatedAt),
	)
	return err
}

func saveReport(e execer, r *input_itf.StepResultEntity) error {
	doc, err := json.Marshal(r.Handoffs)
	if err != nil {
		return err
	}

	usage := ""

	if r.ContextUsage != nil {
		encoded, err := json.Marshal(r.ContextUsage)
		if err != nil {
			return err
		}

		usage = string(encoded)
	}

	_, err = e.Exec(`INSERT OR REPLACE INTO step_results
		(id, step_id, agent_id, attempt_status, handoff, context_usage,
		started_at, completed_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID.String(),
		r.StepID.String(),
		r.AgentID.String(),
		string(r.AttemptStatus),
		string(doc),
		usage,
		formatTime(r.StartedAt),
		formatTime(r.CompletedAt),
		formatTime(r.CreatedAt),
		formatTime(r.UpdatedAt),
	)
	return err
}

func formatTime(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func parseUUID(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}
