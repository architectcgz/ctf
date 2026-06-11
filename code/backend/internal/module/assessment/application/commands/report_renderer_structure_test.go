package commands

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestReportServiceKeepsRendererImplementationsOutOfServiceFile(t *testing.T) {
	t.Parallel()

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, "report_service.go", nil, 0)
	if err != nil {
		t.Fatalf("ParseFile(report_service.go) error = %v", err)
	}

	forbidden := map[string]string{
		"summaryLine":                "shared renderer data belongs in application/reporting",
		"chartRow":                   "PDF chart data belongs in application/reporting",
		"writePersonalPDF":           "standard report PDF writer belongs in application/reporting",
		"writeClassPDF":              "standard report PDF writer belongs in application/reporting",
		"WritePersonalPDF":           "standard report PDF writer belongs in application/reporting",
		"WriteClassPDF":              "standard report PDF writer belongs in application/reporting",
		"writePersonalExcel":         "spreadsheet writer belongs in application/reporting",
		"writeClassExcel":            "spreadsheet writer belongs in application/reporting",
		"WritePersonalExcel":         "spreadsheet writer belongs in application/reporting",
		"WriteClassExcel":            "spreadsheet writer belongs in application/reporting",
		"writeJSONReport":            "JSON writer belongs in application/reporting",
		"WriteJSONReport":            "JSON writer belongs in application/reporting",
		"newReportPDF":               "shared PDF setup belongs in application/reporting",
		"addReportTitle":             "shared PDF helper belongs in application/reporting",
		"addSummaryBlock":            "shared PDF helper belongs in application/reporting",
		"addReportBulletSection":     "shared PDF helper belongs in application/reporting",
		"addReportSectionTitle":      "shared PDF helper belongs in application/reporting",
		"addDimensionChart":          "standard PDF renderer helper belongs in application/reporting",
		"addAverageChart":            "standard PDF renderer helper belongs in application/reporting",
		"addDimensionStatsTable":     "standard PDF renderer helper belongs in application/reporting",
		"addTopStudentsTable":        "standard PDF renderer helper belongs in application/reporting",
		"addClassTrendTable":         "standard PDF renderer helper belongs in application/reporting",
		"addDistributionTable":       "standard PDF renderer helper belongs in application/reporting",
		"addContestMigrationSection": "standard PDF renderer helper belongs in application/reporting",
		"addClassReviewOutlineTable": "standard PDF renderer helper belongs in application/reporting",
		"writePDFTableHeader":        "shared PDF table helper belongs in application/reporting",
		"writePDFCustomTableHeader":  "shared PDF table helper belongs in application/reporting",
		"writePDFTableRow":           "shared PDF table helper belongs in application/reporting",
		"skillProfileChartRows":      "standard PDF renderer helper belongs in application/reporting",
		"ensurePDFSpace":             "shared PDF helper belongs in application/reporting",
		"sanitizePDFText":            "shared PDF helper belongs in application/reporting",
		"localizeReportTerms":        "shared renderer localization belongs in application/reporting",
		"localizeReportTerm":         "shared renderer localization belongs in application/reporting",
		"mustNewExcelStyle":          "spreadsheet helper belongs in application/reporting",
		"writePairs":                 "shared renderer helper belongs in application/reporting",
		"setReportSheetLayout":       "spreadsheet helper belongs in application/reporting",
		"writeDistributionSheet":     "spreadsheet helper belongs in application/reporting",
		"writeReviewSheet":           "spreadsheet helper belongs in application/reporting",
		"writeContestMigrationSheet": "spreadsheet helper belongs in application/reporting",
		"safeTrendPoints":            "shared renderer helper belongs in application/reporting",
		"safeReviewItems":            "shared renderer helper belongs in application/reporting",
		"reviewStudentNames":         "shared renderer helper belongs in application/reporting",
		"reviewSummaryActiveRate":    "shared renderer helper belongs in application/reporting",
		"reviewSummaryRecentEvents":  "shared renderer helper belongs in application/reporting",
	}

	var found []string
	for _, decl := range file.Decls {
		switch typed := decl.(type) {
		case *ast.FuncDecl:
			if reason, ok := forbidden[typed.Name.Name]; ok {
				found = append(found, typed.Name.Name+": "+reason)
			}
		case *ast.GenDecl:
			if typed.Tok != token.TYPE {
				continue
			}
			for _, spec := range typed.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				if reason, ok := forbidden[typeSpec.Name.Name]; ok {
					found = append(found, typeSpec.Name.Name+": "+reason)
				}
			}
		}
	}

	if len(found) > 0 {
		sort.Strings(found)
		t.Fatalf("report_service.go still declares renderer implementation details:\n%s", strings.Join(found, "\n"))
	}
}

func TestReportServiceFileStaysFocused(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("report_service.go")
	if err != nil {
		t.Fatalf("ReadFile(report_service.go) error = %v", err)
	}

	lineCount := strings.Count(string(content), "\n")
	if len(content) > 0 && content[len(content)-1] != '\n' {
		lineCount++
	}

	const maxReportServiceLines = 1000
	if lineCount > maxReportServiceLines {
		t.Fatalf("report_service.go has %d lines, want <= %d; move report data builders and output helpers into focused files", lineCount, maxReportServiceLines)
	}
}

func TestCommandsDoNotOwnReportRenderingAdapters(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob command files: %v", err)
	}

	forbiddenImports := map[string]string{
		"archive/zip":                 "ZIP rendering belongs in application/reporting",
		"github.com/jung-kurt/gofpdf": "PDF rendering belongs in application/reporting",
		"github.com/xuri/excelize/v2": "spreadsheet rendering belongs in application/reporting",
		"ctf-platform/internal/module/assessment/application/reportassets": "PDF assets are consumed by application/reporting",
	}

	fileSet := token.NewFileSet()
	for _, fileName := range files {
		if strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, fileName, nil, parser.ImportsOnly)
		if err != nil {
			t.Fatalf("ParseFile(%s) error = %v", fileName, err)
		}
		for _, importSpec := range file.Imports {
			importPath := strings.Trim(importSpec.Path.Value, `"`)
			if reason, ok := forbiddenImports[importPath]; ok {
				t.Fatalf("%s imports %s: %s", fileName, importPath, reason)
			}
		}
	}
}
