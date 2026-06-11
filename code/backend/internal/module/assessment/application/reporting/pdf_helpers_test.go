package reporting

import "testing"

func TestNewReportPDFRegistersBoldFont(t *testing.T) {
	t.Parallel()

	pdf, err := newReportPDF()
	if err != nil {
		t.Fatalf("newReportPDF() error = %v", err)
	}

	setReportPDFFont(pdf, "B", 14)
	if err := pdf.Error(); err != nil {
		t.Fatalf("expected report pdf bold font to be available, got %v", err)
	}
}
