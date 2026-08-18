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
func validateDraft(template *core_itf.Template) error {
	if strings.TrimSpace(template.Name) == "" {
		return custom_error.Critical("`name` is empty. Give the template a name.")
	}

	if strings.TrimSpace(template.Role) == "" {
		return custom_error.Critical("`role` is empty. Say what an agent built from this template does.")
	}

	if !template.TaskLevel.Valid() {
		return custom_error.Critical(
			"`task_level` is %q, which is not one of: %s.",
			template.TaskLevel, strings.Join(helpers.Labels(enums.TaskLevels()), ", "),
		)
	}

	if err := validatePrompts(template.SystemPrompts); err != nil {
		return err
	}

	return validateParams(template.Params)
}

func validatePrompts(prompts map[string]string) error {
	if len(prompts) == 0 {
		return custom_error.Critical("`system_prompts` is empty. A template needs at least one prompt section.")
	}

	for _, key := range slices.Sorted(maps.Keys(prompts)) {
		if strings.TrimSpace(key) == "" {
			return custom_error.Critical("one entry in `system_prompts` has an empty `key`.")
		}

		value := prompts[key]

		if strings.TrimSpace(value) == "" {
			return custom_error.Critical("the `%s` prompt has no text. Write it or drop the section.", key)
		}

		if len(value) > maxPromptLength {
			return custom_error.Critical(
				"the `%s` prompt is %d characters, over the %d limit. Say the same thing in less.",
				key, len(value), maxPromptLength,
			)
		}
	}

	return nil
}

func validateParams(params map[string]*core_itf.TemplateParams) error {
	if len(params) > maxParams {
		return custom_error.Critical(
			"there are %d params, over the %d limit. Keep only the ones that change what the agent does.",
			len(params), maxParams,
		)
	}

	for _, key := range slices.Sorted(maps.Keys(params)) {
		if strings.TrimSpace(key) == "" {
			return custom_error.Critical("one entry in `params` has an empty `key`.")
		}

		if strings.ContainsAny(key, " \t\n") {
			return custom_error.Critical("the param key `%s` has whitespace in it. Use snake_case.", key)
		}

		param := params[key]

		if param == nil {
			return custom_error.Critical("the param `%s` has no body.", key)
		}

		if strings.TrimSpace(param.Description) == "" {
			return custom_error.Critical("the param `%s` has no `description`.", key)
		}

		if err := validateParamType(key, param); err != nil {
			return err
		}
	}

	return nil
}

func validateParamType(key string, param *core_itf.TemplateParams) error {
	kind := enums.ParamType(param.Type)

	if !kind.Valid() {
		return custom_error.Critical(
			"the param `%s` has type %q, which is not one of: %s.",
			key, param.Type, strings.Join(helpers.Labels(enums.ParamTypes()), ", "),
		)
	}

	if kind != enums.SelectParam && kind != enums.MultiParam {
		return nil
	}

	if len(param.Options) == 0 {
		return custom_error.Critical("the param `%s` is a %s but lists no `options`.", key, kind)
	}

	for _, choice := range defaultChoices(kind, param.Default) {
		if !slices.Contains(param.Options, choice) {
			return custom_error.Critical(
				"the param `%s` defaults to %q, which is not one of its options.", key, choice,
			)
		}
	}

	return nil
}

func defaultChoices(kind enums.ParamType, value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	if kind != enums.MultiParam {
		return []string{value}
	}

	choices := strings.Split(value, ",")
	for i, choice := range choices {
		choices[i] = strings.TrimSpace(choice)
	}

	return choices
}
