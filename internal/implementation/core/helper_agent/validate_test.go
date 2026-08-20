package helper_agent

import (
	"strings"
	"testing"

	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"
)

func goodRole() *core_itf.Role {
	return &core_itf.Role{
		Name:        "Code reviewer",
		Description: "Reads a diff and reports the defects it can prove.",
		Effort:      enums.EffortDeep,
		Inputs: map[string]*core_itf.RoleInputs{
			"diff_path": {Description: "Where the diff is", Type: "text", Required: true},
		},
		Instructions: map[string]string{"base": "Review {{diff_path}}."},
	}
}

func TestValidateDraftAcceptsAFilledRole(t *testing.T) {
	if err := validateDraft(goodRole()); err != nil {
		t.Fatalf("a complete role was refused: %v", err)
	}
}

func TestValidateDraftRefusesWhatTheUserWouldHaveToFixByHand(t *testing.T) {
	cases := []struct {
		name    string
		mangle  func(*core_itf.Role)
		mustSay string
	}{
		{
			name:    "no name",
			mangle:  func(tpl *core_itf.Role) { tpl.Name = "  " },
			mustSay: "name",
		},
		{
			name:    "no role",
			mangle:  func(tpl *core_itf.Role) { tpl.Description = "" },
			mustSay: "description",
		},
		{
			name:    "invented step level",
			mangle:  func(tpl *core_itf.Role) { tpl.Effort = "extremely_heavy" },
			mustSay: "effort",
		},
		{
			name:    "no instructions at all",
			mangle:  func(tpl *core_itf.Role) { tpl.Instructions = nil },
			mustSay: "instructions",
		},
		{
			name: "a prompt with no text",
			mangle: func(tpl *core_itf.Role) {
				tpl.Instructions["limits"] = "   "
			},
			mustSay: "limits",
		},
		{
			name: "a input with no description",
			mangle: func(tpl *core_itf.Role) {
				tpl.Inputs["diff_path"].Description = ""
			},
			mustSay: "description",
		},
		{
			name: "a input with an invented type",
			mangle: func(tpl *core_itf.Role) {
				tpl.Inputs["diff_path"].Type = "freeform"
			},
			mustSay: "diff_path",
		},
		{
			name: "a select with no options",
			mangle: func(tpl *core_itf.Role) {
				tpl.Inputs["diff_path"].Type = string(enums.SelectInput)
			},
			mustSay: "options",
		},
		{
			name: "a select defaulting outside its options",
			mangle: func(tpl *core_itf.Role) {
				tpl.Inputs["diff_path"].Type = string(enums.SelectInput)
				tpl.Inputs["diff_path"].Options = []string{"staged", "head"}
				tpl.Inputs["diff_path"].Default = "branch"
			},
			mustSay: "options",
		},
		{
			name: "a input key with whitespace in it",
			mangle: func(tpl *core_itf.Role) {
				tpl.Inputs["diff path"] = tpl.Inputs["diff_path"]
				delete(tpl.Inputs, "diff_path")
			},
			mustSay: "snake_case",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			role := goodRole()
			testCase.mangle(role)

			err := validateDraft(role)
			if err == nil {
				t.Fatal("the role was accepted")
			}

			if !strings.Contains(err.Error(), testCase.mustSay) {
				t.Errorf("refusal %q does not name %q, so the agent cannot fix it", err, testCase.mustSay)
			}
		})
	}
}

func TestValidateDraftRefusesMoreInputsThanAnyoneWouldFillIn(t *testing.T) {
	role := goodRole()

	for _, key := range []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m"} {
		role.Inputs[key] = &core_itf.RoleInputs{Description: "filler", Type: "text"}
	}

	if err := validateDraft(role); err == nil {
		t.Error("a role with more inputs than the limit was accepted")
	}
}

func TestValidateDraftAcceptsARoleWithNoInputs(t *testing.T) {
	role := goodRole()
	role.Inputs = nil

	if err := validateDraft(role); err != nil {
		t.Errorf("a role that needs no inputs was refused: %v", err)
	}
}
