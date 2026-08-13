package helper_agent

import (
	"strings"
	"testing"

	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"
)

func goodTemplate() *core_itf.Template {
	return &core_itf.Template{
		Name:      "Code reviewer",
		Role:      "Reads a diff and reports the defects it can prove.",
		TaskLevel: enums.HeavyTask,
		Params: map[string]*core_itf.TemplateParams{
			"diff_path": {Description: "Where the diff is", Type: "text", Required: true},
		},
		SystemPrompts: map[string]string{"base": "Review {{diff_path}}."},
	}
}

func TestValidateDraftAcceptsAFilledTemplate(t *testing.T) {
	if err := validateDraft(goodTemplate()); err != nil {
		t.Fatalf("a complete template was refused: %v", err)
	}
}

func TestValidateDraftRefusesWhatTheUserWouldHaveToFixByHand(t *testing.T) {
	cases := []struct {
		name    string
		mangle  func(*core_itf.Template)
		mustSay string
	}{
		{
			name:    "no name",
			mangle:  func(tpl *core_itf.Template) { tpl.Name = "  " },
			mustSay: "name",
		},
		{
			name:    "no role",
			mangle:  func(tpl *core_itf.Template) { tpl.Role = "" },
			mustSay: "role",
		},
		{
			name:    "invented task level",
			mangle:  func(tpl *core_itf.Template) { tpl.TaskLevel = "extremely_heavy" },
			mustSay: "task_level",
		},
		{
			name:    "no prompts at all",
			mangle:  func(tpl *core_itf.Template) { tpl.SystemPrompts = nil },
			mustSay: "system_prompts",
		},
		{
			name: "a prompt with no text",
			mangle: func(tpl *core_itf.Template) {
				tpl.SystemPrompts["limits"] = "   "
			},
			mustSay: "limits",
		},
		{
			name: "a param with no description",
			mangle: func(tpl *core_itf.Template) {
				tpl.Params["diff_path"].Description = ""
			},
			mustSay: "description",
		},
		{
			name: "a param with an invented type",
			mangle: func(tpl *core_itf.Template) {
				tpl.Params["diff_path"].Type = "freeform"
			},
			mustSay: "diff_path",
		},
		{
			name: "a select with no options",
			mangle: func(tpl *core_itf.Template) {
				tpl.Params["diff_path"].Type = string(enums.SelectParam)
			},
			mustSay: "options",
		},
		{
			name: "a select defaulting outside its options",
			mangle: func(tpl *core_itf.Template) {
				tpl.Params["diff_path"].Type = string(enums.SelectParam)
				tpl.Params["diff_path"].Options = []string{"staged", "head"}
				tpl.Params["diff_path"].Default = "branch"
			},
			mustSay: "options",
		},
		{
			name: "a param key with whitespace in it",
			mangle: func(tpl *core_itf.Template) {
				tpl.Params["diff path"] = tpl.Params["diff_path"]
				delete(tpl.Params, "diff_path")
			},
			mustSay: "snake_case",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			template := goodTemplate()
			testCase.mangle(template)

			err := validateDraft(template)
			if err == nil {
				t.Fatal("the template was accepted")
			}

			if !strings.Contains(err.Error(), testCase.mustSay) {
				t.Errorf("refusal %q does not name %q, so the agent cannot fix it", err, testCase.mustSay)
			}
		})
	}
}

func TestValidateDraftRefusesMoreParamsThanAnyoneWouldFillIn(t *testing.T) {
	template := goodTemplate()

	for _, key := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m"} {
		template.Params[key] = &core_itf.TemplateParams{Description: "filler", Type: "text"}
	}

	if err := validateDraft(template); err == nil {
		t.Error("a template with more params than the limit was accepted")
	}
}

func TestValidateDraftAcceptsATemplateWithNoParams(t *testing.T) {
	template := goodTemplate()
	template.Params = nil

	if err := validateDraft(template); err != nil {
		t.Errorf("a template that needs no inputs was refused: %v", err)
	}
}
