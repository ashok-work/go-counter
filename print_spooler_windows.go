//go:build windows

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/phpdave11/gofpdf"
)

func getPrinters() ([]string, error) {
	script := `
$payload = [ordered]@{
	printers = @()
}

$names = @()

if (Get-Command Get-Printer -ErrorAction SilentlyContinue) {
	try {
		$names += Get-Printer -ErrorAction Stop | Select-Object -ExpandProperty Name
	} catch {
	}
}

if (-not $names -and (Get-Command Get-CimInstance -ErrorAction SilentlyContinue)) {
	try {
		$names += Get-CimInstance Win32_Printer -ErrorAction Stop | Select-Object -ExpandProperty Name
	} catch {
	}
}

if (-not $names -and (Get-Command Get-WmiObject -ErrorAction SilentlyContinue)) {
	try {
		$names += Get-WmiObject Win32_Printer -ErrorAction Stop | Select-Object -ExpandProperty Name
	} catch {
	}
}

$payload.printers = @(
	$names |
	ForEach-Object { "$_".Trim() } |
	Where-Object { $_ } |
	Sort-Object -Unique
)

$payload | ConvertTo-Json -Compress
`
	out, err := runPowerShell(script)
	if err != nil {
		return nil, fmt.Errorf("failed to list printers: %w", err)
	}

	var payload struct {
		Printers []string `json:"printers"`
	}

	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		log.Printf("windows printer discovery returned no output")
		return []string{}, nil
	}

	if err := json.Unmarshal([]byte(trimmed), &payload); err != nil {
		return nil, fmt.Errorf("failed to parse printer list: %w: %s", err, trimmed)
	}

	log.Printf("windows printer discovery found %d printers", len(payload.Printers))
	return payload.Printers, nil
}

func submitPDF(pdf *gofpdf.Fpdf, filename string) error {
	dir, err := os.MkdirTemp("", "fyne-browser-print-*")
	if err != nil {
		return err
	}

	filePath := filepath.Join(dir, filename)
	if err := pdf.OutputFileAndClose(filePath); err != nil {
		return err
	}

	script := buildWindowsPrintScript(filePath, getSelectedPrinter())
	out, err := runPowerShell(script)
	if err != nil {
		return fmt.Errorf("windows print failed for %s: %w", filePath, err)
	}

	log.Printf("print job submitted: %s", strings.TrimSpace(string(out)))
	_ = os.Remove(filePath)
	_ = os.Remove(dir)
	return nil
}

func buildWindowsPrintScript(filePath, printer string) string {
	quotedFile := powershellQuote(filePath)
	if strings.TrimSpace(printer) == "" {
		return fmt.Sprintf("Start-Process -FilePath %s -Verb Print; Start-Sleep -Seconds 2", quotedFile)
	}

	quotedPrinter := powershellQuote(printer)
	return fmt.Sprintf("Start-Process -FilePath %s -Verb PrintTo -ArgumentList %s; Start-Sleep -Seconds 2", quotedFile, quotedPrinter)
}

func powershellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func runPowerShell(script string) ([]byte, error) {
	powershellPath, err := resolvePowerShell()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(powershellPath, "-NoProfile", "-NonInteractive", "-Command", script)
	out, runErr := cmd.CombinedOutput()
	if runErr != nil {
		return out, fmt.Errorf("%w: %s", runErr, strings.TrimSpace(string(out)))
	}

	return out, nil
}

func resolvePowerShell() (string, error) {
	candidates := []string{
		"powershell.exe",
		"powershell",
		"pwsh.exe",
		"pwsh",
	}

	systemRoots := []string{
		strings.TrimSpace(os.Getenv("SystemRoot")),
		strings.TrimSpace(os.Getenv("WINDIR")),
		`C:\Windows`,
	}

	for _, root := range systemRoots {
		if root == "" {
			continue
		}
		candidates = append(candidates,
			filepath.Join(root, "System32", "WindowsPowerShell", "v1.0", "powershell.exe"),
			filepath.Join(root, "Sysnative", "WindowsPowerShell", "v1.0", "powershell.exe"),
			filepath.Join(root, "System32", "WindowsPowerShell", "v1.0", "powershell"),
			filepath.Join(root, "System32", "WindowsPowerShell", "v1.0", "pwsh.exe"),
			filepath.Join(root, "System32", "WindowsPowerShell", "v1.0", "pwsh"),
			filepath.Join(root, "System32", "pwsh.exe"),
			filepath.Join(root, "System32", "pwsh"),
		)
	}

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}

		if filepath.IsAbs(candidate) {
			if fileInfo, statErr := os.Stat(candidate); statErr == nil && !fileInfo.IsDir() {
				return candidate, nil
			}
			continue
		}

		resolved, lookErr := exec.LookPath(candidate)
		if lookErr == nil {
			return resolved, nil
		}
	}

	return "", fmt.Errorf("powershell executable not found; checked PATH and standard Windows locations")
}
