package helper_agent

import (
	"maps"
	"slices"
	"strings"

	"hexago/internal/helpers"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"
)

// validateDraft is the reason the tool call is worth having: the schema already
// rejected the wrong shape, and this rejects the wrong content. Every message names
// the field and what to do about it, because the reader is the agent that has to fix
// it and call again.
func validateDraft(role *core_itf.Role) error {
	if strings.TrimSpace(role.Name) == "" {
		return custom_error.Critical("`name` is empty. Give the role a name.")
	}

	if strings.TrimSpace(role.Description) == "" {
		return custom_error.Critical("`description` is empty. Say what an agent built from this role does.")
	}

	if !role.Effort.Valid() {
		return custom_error.Critical(
			"`effort` is %q, which is not one of: %s.",
			role.Effort, strings.Join(helpers.Labels(enums.Efforts()), ", "),
		)
	}

	if err := validatePrompts(role.Instructions); err != nil {
		return err
	}

	return validateInputs(role.Inputs)
}

func validatePrompts(prompts map[string]string) error {
	if len(prompts) == 0 {
		return custom_error.Critical("`instructions` is empty. A role needs at least one instruction section.")
	}

	for _, key := range slices.Sorted(maps.Keys(prompts)) {
		if strings.TrimSpace(key) == "" {
			return custom_error.Critical("one entry in `instructions` has an empty `key`.")
		}

		value := prompts[key]

		if strings.TrimSpace(value) == "" {
			return custom_error.Critical("the `%s` instruction has no text. Write it or drop the section.", key)
		}

		if len(value) > maxPromptLength {
			return custom_error.Critical(
				"the `%s` instruction is %d characters, over the %d limit. Say the same thing in less.",
				key, len(value), maxPromptLength,
			)
		}
	}

	return nil
}

func validateInputs(inputs map[string]*core_itf.RoleInputs) error {
	if len(inputs) > maxInputs {
		return custom_error.Critical(
			"there are %d inputs, over the %d limit. Keep only the ones that change what the agent does.",
			len(inputs), maxInputs,
		)
	}

	for _, key := range slices.Sorted(maps.Keys(inputs)) {
		if strings.TrimSpace(key) == "" {
			return custom_error.Critical("one entry in `inputs` has an empty `key`.")
		}

		if strings.ContainsAny(key, " \t\n") {
			return custom_error.Critical("the input key `%s` has whitespace in it. Use snake_case.", key)
		}

		input := inputs[key]

		if input == nil {
			return custom_error.Critical("the input `%s` has no body.", key)
		}

		if strings.TrimSpace(input.Description) == "" {
			return custom_error.Critical("the input `%s` has no `description`.", key)
		}

		if err := validateInputType(key, input); err != nil {
			return err
		}
	}

	return nil
}

func validateInputType(key string, input *core_itf.RoleInputs) error {
	kind := enums.InputType(input.Type)

	if !kind.Valid() {
		return custom_error.Critical(
			"the input `%s` has type %q, which is not one of: %s.",
			key, input.Type, strings.Join(helpers.Labels(enums.InputTypes()), ", "),
		)
	}

	if kind != enums.SelectInput && kind != enums.MultiInput {
		return nil
	}

	if len(input.Options) == 0 {
		return custom_error.Critical("the input `%s` is a %s but lists no `options`.", key, kind)
	}

	for _, choice := range defaultChoices(kind, input.Default) {
		if !slices.Contains(input.Options, choice) {
			return custom_error.Critical(
				"the input `%s` defaults to %q, which is not one of its options.", key, choice,
			)
		}
	}

	return nil
}

func defaultChoices(kind enums.InputType, value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	if kind != enums.MultiInput {
		return []string{value}
	}

	choices := strings.Split(value, ",")
	for i, choice := range choices {
		choices[i] = strings.TrimSpace(choice)
	}

	return choices
}
