package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/DuckInAShirt/leetmate/internal/config"
	"github.com/DuckInAShirt/leetmate/internal/doctor"
	"github.com/DuckInAShirt/leetmate/internal/tui"
)

var errDoctorFailed = errors.New("environment check failed")

func runDoctor(args []string, out io.Writer) error {
	fs := flag.NewFlagSet("doctor", flag.ContinueOnError)
	fs.SetOutput(out)
	asJSON := fs.Bool("json", false, "print machine-readable JSON")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected args: %v", fs.Args())
	}

	cfg, configPath, err := config.Load()
	if err != nil {
		report := doctor.Report{Checks: []doctor.Check{{ID: "config", Level: doctor.Fail, Reason: "unreadable", Value: configPath, Extra: err.Error()}}}
		if *asJSON {
			enc := json.NewEncoder(out)
			enc.SetIndent("", "  ")
			if encodeErr := enc.Encode(report); encodeErr != nil {
				return encodeErr
			}
		} else {
			printDoctorReport(out, report, cfg.Language)
		}
		return errDoctorFailed
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	report := doctor.Run(cfg, configPath, cwd)
	if *asJSON {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			return err
		}
	} else {
		printDoctorReport(out, report, cfg.Language)
	}
	if report.HasFailures() {
		return errDoctorFailed
	}
	return nil
}

func printDoctorReport(out io.Writer, report doctor.Report, language string) {
	fmt.Fprintln(out, tui.Text(language, "doctor.title"))
	for _, check := range report.Checks {
		fmt.Fprintf(out, "%s %-10s %s\n", doctorMark(check.Level), tui.Text(language, "doctor.label."+check.ID), doctorMessage(check, language))
	}
	if report.HasFailures() {
		fmt.Fprintln(out, tui.Text(language, "doctor.next.fix"))
	} else {
		fmt.Fprintln(out, tui.Text(language, "doctor.ready"))
	}
}

func doctorMark(level doctor.Level) string {
	switch level {
	case doctor.Pass:
		return "[PASS]"
	case doctor.Warn:
		return "[WARN]"
	default:
		return "[FAIL]"
	}
}

func doctorMessage(check doctor.Check, language string) string {
	key := "doctor." + check.ID + "." + check.Reason
	switch key {
	case "doctor.config.found", "doctor.leetgo.found", "doctor.workspace.ready", "doctor.auth.runtime_unverified", "doctor.llm.found", "doctor.llm.missing", "doctor.llm.invalid_provider":
		return tui.Textf(language, key, check.Value)
	case "doctor.config.unreadable", "doctor.workspace.invalid_config", "doctor.workspace.not_directory", "doctor.workspace.no_config":
		value := check.Extra
		if value == "" {
			value = check.Value
		}
		return tui.Textf(language, key, value)
	case "doctor.config_dir.writable", "doctor.data.writable":
		return tui.Textf(language, "doctor.path.writable", check.Value)
	case "doctor.config_dir.unwritable", "doctor.data.unwritable":
		return tui.Textf(language, "doctor.path.unwritable", check.Value, check.Extra)
	default:
		translated := tui.Text(language, key)
		if translated != key {
			return translated
		}
		if check.Extra != "" {
			return check.Reason + ": " + check.Extra
		}
		return check.Reason
	}
}
