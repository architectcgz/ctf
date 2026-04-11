package jobs

type awdHTTPCheckerActionRuntimeResult struct {
	summary      *awdCheckActionResult
	responseBody string
}

type awdHTTPCheckerTargetRuntimeResult struct {
	status       string
	statusReason string
	target       awdCheckTargetResult
}
