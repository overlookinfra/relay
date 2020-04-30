package error

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/puppetlabs/errawr-go/v2/pkg/encoding"
	"github.com/puppetlabs/relay/pkg/config"
	"github.com/puppetlabs/relay/pkg/dialog"
	"github.com/puppetlabs/relay/pkg/errors"
	"github.com/spf13/cobra"
)

func FormatError(err error, cmd *cobra.Command) {
	// attempt to load config for display options.
	cfg, cfgerr := config.GetConfig(cmd.Flags())

	// if there was a problem loading config use default config
	if cfgerr != nil {
		cfg = config.GetDefaultConfig()
	}

	if cfg.Out == config.OutputTypeJSON {
		formatJSONError(coerceErrawr(err))
	} else {
		formatTextError(coerceErrawr(err), cfg)
	}
}

// coerceErrawr ensures all errors come from errors.yaml as a last-ditch effort
func coerceErrawr(err error) errors.Error {
	errawr, ok := err.(errors.Error)

	if ok {
		return errawr
	}

	return errors.NewGeneralUnknownError().WithCause(err)
}

// formatJSONError uses errawr envelope encoding to generate a json display of an error
// We could make a condensed json representation but it is very useful to use
// the one we already have for now
func formatJSONError(err errors.Error) {
	display := encoding.ForDisplay(err)
	jsonBytes, _ := json.MarshalIndent(display, "", "  ")

	fmt.Println(string(jsonBytes))
}

func formatTextError(err errors.Error, cfg *config.Config) {
	log := dialog.NewDialog(cfg)

	var out string

	appendError(err, cfg, &out, 0, "")

	if cfg.Debug {
		out += fmt.Sprintf(`

You have recieved an error in debug mode. If the error persists you may file a bug report at https://github.com/puppetlabs/relay/issues`)
	}

	log.Error(out)
}

// appendError recursively prints errawr causes and items, progressively indented
func appendError(err errors.Error, cfg *config.Config, out *string, indent int, prefix string) {
	// print error if in debug mode or if Sensitivity is zero
	if err.Sensitivity() == 0 || cfg.Debug {
		*out += strings.Repeat(" ", indent)

		if prefix != "" {
			*out += prefix
		}

		*out += err.FormattedDescription().Friendly()

		for _, cause := range err.Causes() {
			*out += "\n"
			appendError(cause, cfg, out, indent+2, "• ")
		}

		if items, ok := err.Items(); ok {
			for itemKey, item := range items {
				*out += "\n"
				appendError(item, cfg, out, indent+2, fmt.Sprintf("• `%v`", itemKey))
			}
		}
	}
}