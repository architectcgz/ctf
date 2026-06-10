import { request } from '../request'
import type { ContestAnnouncementSyncData } from '../contracts'

import type {
  AdminContestAnnouncementCreatePayload,
  RawContestAnnouncement,
  ContestAnnouncement,
} from './contest-support'

// Re-export types for barrel consumers
export type { AdminContestAnnouncementCreatePayload } from './contest-support'

import { normalizeContestAnnouncement } from './contest-support'

interface RawContestAnnouncementSyncEvent {
  cursor: string | number
  type: ContestAnnouncementSyncData['events'][number]['type']
  announcement?: RawContestAnnouncement
  announcement_id?: string | number
  occurred_at: string
}

interface RawContestAnnouncementSyncData {
  events: RawContestAnnouncementSyncEvent[]
  next_cursor: string | number
  has_more: boolean
}

function normalizeContestAnnouncementSync(
  data: RawContestAnnouncementSyncData
): ContestAnnouncementSyncData {
  return {
    events: data.events.map((event) => ({
      cursor: String(event.cursor),
      type: event.type,
      announcement: event.announcement
        ? normalizeContestAnnouncement(event.announcement)
        : undefined,
      announcement_id: event.announcement_id == null ? undefined : String(event.announcement_id),
      occurred_at: event.occurred_at,
    })),
    next_cursor: String(data.next_cursor),
    has_more: data.has_more,
  }
}

export async function getAdminContestAnnouncements(
  contestId: string
): Promise<ContestAnnouncement[]> {
  const response = await request<RawContestAnnouncement[]>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/announcements`,
  })

  return response.map(normalizeContestAnnouncement)
}

export async function createAdminContestAnnouncement(
  contestId: string,
  data: AdminContestAnnouncementCreatePayload
): Promise<ContestAnnouncement> {
  const response = await request<RawContestAnnouncement>({
    method: 'POST',
    url: `/admin/contests/${encodeURIComponent(contestId)}/announcements`,
    data: {
      title: data.title,
      content: data.content,
    },
  })

  return normalizeContestAnnouncement(response)
}

export async function deleteAdminContestAnnouncement(
  contestId: string,
  announcementId: string
): Promise<void> {
  await request<void>({
    method: 'DELETE',
    url: `/admin/contests/${encodeURIComponent(contestId)}/announcements/${encodeURIComponent(announcementId)}`,
  })
}

export async function getAdminContestAnnouncementSync(
  contestId: string,
  afterId?: string
): Promise<ContestAnnouncementSyncData> {
  const response = await request<RawContestAnnouncementSyncData>({
    method: 'GET',
    url: `/admin/contests/${encodeURIComponent(contestId)}/announcements/sync`,
    params: afterId ? { after_id: afterId } : undefined,
  })

  return normalizeContestAnnouncementSync(response)
}
