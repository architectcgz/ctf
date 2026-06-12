package commands

import opscontracts "ctf-platform/internal/module/ops/contracts"

const (
	NotificationAudienceTypeAll   = "all"
	NotificationAudienceTypeRole  = "role"
	NotificationAudienceTypeClass = "class"
	NotificationAudienceTypeUser  = "user"
)

type NotificationInfo = opscontracts.NotificationInfo

type AdminNotificationPublishResp struct {
	BatchID        int64 `json:"batch_id"`
	RecipientCount int   `json:"recipient_count"`
}
