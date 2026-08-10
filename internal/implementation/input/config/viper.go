package config

import (
	"github.com/spf13/viper"

	"hexago/internal/helpers"
	"hexago/internal/helpers/constances"
	"hexago/internal/helpers/custom_error"
	input_itf "hexago/internal/interface/input"
)

type viperConfig struct {
	cfg *input_itf.ConfigStruct
}

func New(path string) (input_itf.Config, error) {
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	cfg := &input_itf.ConfigStruct{}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, custom_error.Critical("cannot decode %s: %v", path, err)
	}

	if err := helpers.ValidateStruct(cfg); err != nil {
		return nil, custom_error.Critical("invalid %s: %v", path, err)
	}

	if err := validatePKCE(cfg.MCPServers); err != nil {
		return nil, custom_error.Critical("invalid %s: %v", path, err)
	}

	return &viperConfig{cfg: cfg}, nil
}

func validatePKCE(cfg *input_itf.MCPServersConfig) error {
	if cfg.VerifierBytes < constances.PKCEMinVerifierBytes {
		return custom_error.Critical(
			"mcp_servers.verifier_bytes is %d, must be at least %d (RFC 7636 floor)",
			cfg.VerifierBytes, constances.PKCEMinVerifierBytes,
		)
	}

	if cfg.StateBytes < constances.PKCEMinStateBytes {
		return custom_error.Critical(
			"mcp_servers.state_bytes is %d, must be at least %d",
			cfg.StateBytes, constances.PKCEMinStateBytes,
		)
	}

	if cfg.ChallengeMethod != constances.PKCESupportedChallengeMethod {
		return custom_error.Critical(
			"mcp_servers.challenge_method is %q, the only supported method is %q",
			cfg.ChallengeMethod, constances.PKCESupportedChallengeMethod,
		)
	}

	return nil
}

func (c *viperConfig) Read() *input_itf.ConfigStruct {
	return c.cfg
}
