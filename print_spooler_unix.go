//go:build darwin || linux

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/phpdave11/gofpdf"
)

func getPrinters() ([]string, error) {
	cmd := exec.Command("lpstat", "-e")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list printers: %w: %s", err, strings.TrimSpace(string(out)))
	}

	seen := map[string]struct{}{}
	printers := make([]string, 0)
	for _, entry := range strings.Fields(string(out)) {
		name := strings.TrimSpace(entry)
		if name == "" {
			continue
		}
		if _, exists := seen[name]; exists {
			continue
		}
		seen[name] = struct{}{}
		printers = append(printers, name)
	}

	return printers, nil
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

	args := make([]string, 0, 3)
	if printer := getSelectedPrinter(); printer != "" {
		args = append(args, "-d", printer)
	}
	args = append(args, filePath)

	cmd := exec.Command("lp", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("lp failed for %s: %w: %s", filePath, err, strings.TrimSpace(string(out)))
	}

	log.Printf("print job submitted: %s", strings.TrimSpace(string(out)))
	_ = os.Remove(filePath)
	_ = os.Remove(dir)
	return nil
}
