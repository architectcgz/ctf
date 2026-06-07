import type {
  ContestAnnouncement,
  ContestAnnouncementSyncData,
  ContestAnnouncementSyncEvent,
} from '@/api/contracts'

function compareAnnouncementsDesc(left: ContestAnnouncement, right: ContestAnnouncement): number {
  const createdAtDiff = new Date(right.created_at).getTime() - new Date(left.created_at).getTime()
  if (createdAtDiff !== 0) {
    return createdAtDiff
  }

  return String(right.id).localeCompare(String(left.id), 'en')
}

function upsertAnnouncement(
  announcements: ContestAnnouncement[],
  nextAnnouncement: ContestAnnouncement
): ContestAnnouncement[] {
  const filtered = announcements.filter((item) => item.id !== nextAnnouncement.id)
  filtered.push(nextAnnouncement)
  filtered.sort(compareAnnouncementsDesc)
  return filtered
}

export function applyContestAnnouncementSyncEvents(
  announcements: ContestAnnouncement[],
  events: ContestAnnouncementSyncEvent[]
): ContestAnnouncement[] {
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

export function nextContestAnnouncementSyncCursor(sync: ContestAnnouncementSyncData): string {
  return String(sync.next_cursor || '')
}
