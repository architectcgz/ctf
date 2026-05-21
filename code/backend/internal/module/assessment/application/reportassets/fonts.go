package reportassets

import _ "embed"

//go:embed fonts/NotoSansSC-Regular.ttf
var reportPDFFontRegular []byte

//go:embed fonts/NotoSansSC-Bold.ttf
var reportPDFFontBold []byte

func ReportPDFRegularFont() []byte {
	return reportPDFFontRegular
}

func ReportPDFBoldFont() []byte {
	return reportPDFFontBold
}
