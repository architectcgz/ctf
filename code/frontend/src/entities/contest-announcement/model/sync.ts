export interface ContestAnnouncementSyncItem {
  id: string
  created_at: string
}

export type ContestAnnouncementSyncEventType =
  | 'contest.announcement.created'
  | 'contest.announcement.deleted'

export interface ContestAnnouncementSyncEventSource<
  TAnnouncement extends ContestAnnouncementSyncItem,
> {
  type: ContestAnnouncementSyncEventType
  announcement?: TAnnouncement
  announcement_id?: string
}

export interface ContestAnnouncementSyncCursorSource {
  next_cursor?: string | null
}

function compareAnnouncementsDesc<TAnnouncement extends ContestAnnouncementSyncItem>(
  left: TAnnouncement,
  right: TAnnouncement
): number {
  const createdAtDiff = new Date(right.created_at).getTime() - new Date(left.created_at).getTime()
  if (createdAtDiff !== 0) {
    return createdAtDiff
  }

  return String(right.id).localeCompare(String(left.id), 'en')
}

function upsertAnnouncement<TAnnouncement extends ContestAnnouncementSyncItem>(
  announcements: TAnnouncement[],
  nextAnnouncement: TAnnouncement
): TAnnouncement[] {
  const filtered = announcements.filter((item) => item.id !== nextAnnouncement.id)
  filtered.push(nextAnnouncement)
  filtered.sort(compareAnnouncementsDesc)
  return filtered
}

export function applyContestAnnouncementSyncEvents<
  TAnnouncement extends ContestAnnouncementSyncItem,
>(
  announcements: TAnnouncement[],
  events: ContestAnnouncementSyncEventSource<TAnnouncement>[]
): TAnnouncement[] {
  let nextAnnouncements = announcements.slice()

  for (const event of events) {
    if (event.type === 'contest.announcement.created' && event.announcement) {
      nextAnnouncements = upsertAnnouncement(nextAnnouncements, event.announcement)
      continue
    }

    if (event.type === 'contest.announcement.deleted' && event.announcement_id) {
      nextAnnouncements = nextAnnouncements.filter((item) => item.id !== event.announcement_id)
    }
  }

  return nextAnnouncements
}

export function nextContestAnnouncementSyncCursor(
  sync: ContestAnnouncementSyncCursorSource
): string {
  return String(sync.next_cursor || '')
}
