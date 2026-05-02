package commands

type SendNotificationInput struct {
	Type    string
	Title   string
	Content string
	Link    *string
}

type NotificationAudienceRuleInput struct {
	Type   string
	Values []string
}

type NotificationAudienceRulesInput struct {
	Mode  string
	Rules []NotificationAudienceRuleInput
}

type PublishAdminNotificationInput struct {
	Type          string
	Title         string
	Content       string
	Link          *string
	AudienceRules NotificationAudienceRulesInput
}
