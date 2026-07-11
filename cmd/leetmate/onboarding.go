package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/DuckInAShirt/leetmate/internal/config"
	"github.com/DuckInAShirt/leetmate/internal/doctor"
	"github.com/DuckInAShirt/leetmate/internal/tui"
)

func runFirstRun(in io.Reader, out io.Writer, cwd string, interactive bool) (bool, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return false, err
	}
	configPath := filepath.Join(dir, "config.yaml")
	if _, err := os.Stat(configPath); err == nil {
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	workspace, err := doctor.FindWorkspace(cwd)
	if err != nil {
		return false, err
	}
	if !interactive {
		fmt.Fprintln(out, tui.Text("en", "onboarding.not_configured"))
		if workspace != "" {
			fmt.Fprintln(out, tui.Textf("en", "onboarding.run_init", workspace))
		} else {
			fmt.Fprintln(out, tui.Text("en", "onboarding.run_leetgo_init"))
		}
		return true, errDoctorFailed
	}

	reader := bufio.NewReader(in)
	lang := "zh"
	fmt.Fprintln(out, tui.Text(lang, "onboarding.title"))
	lang, err = promptChoice(reader, out, tui.Text(lang, "onboarding.language"), "zh", map[string]bool{"zh": true, "en": true}, lang)
	if err != nil {
		return true, err
	}
	if workspace == "" {
		workspace, err = promptWorkspace(reader, out, cwd, lang)
		if err != nil {
			return true, err
		}
	} else {
		fmt.Fprintln(out, tui.Textf(lang, "onboarding.found_workspace", workspace))
	}
	preset := "siliconflow"
	if lang == "en" {
		preset = "gemini"
	}
	if err := runInit([]string{"--lang", lang, "--workspace", workspace, "--preset", preset}, out); err != nil {
		return true, err
	}
	fmt.Fprintln(out, tui.Text(lang, "onboarding.environment_check"))
	if err := runDoctor(nil, out); err != nil {
		return true, err
	}
	fmt.Fprintln(out, tui.Text(lang, "onboarding.continue"))
	if _, err := reader.ReadString('\n'); err != nil && !errors.Is(err, io.EOF) {
		return true, err
	}
	return true, nil
}

func promptChoice(reader *bufio.Reader, out io.Writer, prompt, fallback string, valid map[string]bool, language string) (string, error) {
	for {
		fmt.Fprint(out, prompt)
		value, err := readLine(reader)
		if err != nil {
			return "", err
		}
		if value == "" {
			value = fallback
		}
		if valid[value] {
			return value, nil
		}
		fmt.Fprintln(out, tui.Text(language, "onboarding.invalid_choice"))
	}
}

func promptWorkspace(reader *bufio.Reader, out io.Writer, cwd, language string) (string, error) {
	for {
		fmt.Fprint(out, tui.Textf(language, "onboarding.workspace", cwd))
		value, err := readLine(reader)
		if err != nil {
			return "", err
		}
		if value == "" {
			value = cwd
		}
		value, err = filepath.Abs(value)
		if err != nil {
			fmt.Fprintln(out, tui.Text(language, "onboarding.invalid_path"))
			continue
		}
		if info, err := os.Stat(filepath.Join(value, "leetgo.yaml")); err == nil && !info.IsDir() {
			return value, nil
		}
		fmt.Fprintln(out, tui.Text(language, "onboarding.no_workspace"))
	}
}

func readLine(reader *bufio.Reader) (string, error) {
	value, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	if errors.Is(err, io.EOF) && value == "" {
		return "", io.EOF
	}
	return strings.TrimSpace(value), nil
}
