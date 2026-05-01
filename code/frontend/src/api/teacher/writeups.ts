import { request } from '../request'

import type {
  PageResult,
  SubmissionWriteupData,
  TeacherManualReviewSubmissionDetailData,
  TeacherManualReviewSubmissionItemData,
  TeacherSubmissionWriteupItemData,
} from '../contracts'

interface RawTeacherSubmissionWriteupItem extends Omit<
  TeacherSubmissionWriteupItemData,
  'id' | 'user_id' | 'challenge_id'
> {
  id: string | number
  user_id: string | number
  challenge_id: string | number
}

interface RawTeacherManualReviewSubmissionItem extends Omit<
  TeacherManualReviewSubmissionItemData,
  'id' | 'user_id' | 'challenge_id'
> {
  id: string | number
  user_id: string | number
  challenge_id: string | number
}

interface RawTeacherManualReviewSubmissionDetail extends Omit<
  TeacherManualReviewSubmissionDetailData,
  'id' | 'user_id' | 'challenge_id'
> {
  id: string | number
  user_id: string | number
  challenge_id: string | number
}

function normalizeSubmissionWriteupData(
  item: SubmissionWriteupData & {
    id: string | number
    user_id: string | number
    challenge_id: string | number
    contest_id?: string | number
    recommended_by?: string | number
  }
): SubmissionWriteupData {
  return {
    ...item,
    id: String(item.id),
    user_id: String(item.user_id),
    challenge_id: String(item.challenge_id),
    contest_id: item.contest_id == null ? undefined : String(item.contest_id),
    recommended_by: item.recommended_by == null ? undefined : String(item.recommended_by),
  }
}

export async function getTeacherWriteupSubmissions(params?: {
  student_id?: string
  challenge_id?: string
  class_name?: string
  submission_status?: 'draft' | 'published'
  visibility_status?: 'visible' | 'hidden'
  page?: number
  page_size?: number
}): Promise<PageResult<TeacherSubmissionWriteupItemData>> {
  const payload = await request<PageResult<RawTeacherSubmissionWriteupItem>>({
    method: 'GET',
    url: '/teacher/writeup-submissions',
    params: {
      student_id: params?.student_id,
      challenge_id: params?.challenge_id,
      class_name: params?.class_name,
      submission_status: params?.submission_status,
      visibility_status: params?.visibility_status,
      page: params?.page,
      page_size: params?.page_size,
    },
  })

  return {
    ...payload,
    list: payload.list.map((item) => ({
      ...item,
      id: String(item.id),
      user_id: String(item.user_id),
      challenge_id: String(item.challenge_id),
    })),
  }
}

export async function recommendTeacherCommunityWriteup(id: string): Promise<SubmissionWriteupData> {
  const payload = await request<
    SubmissionWriteupData & {
      id: string | number
      user_id: string | number
      challenge_id: string | number
      contest_id?: string | number
      recommended_by?: string | number
    }
  >({
    method: 'POST',
    url: `/teacher/community-writeups/${encodeURIComponent(id)}/recommend`,
  })
  return normalizeSubmissionWriteupData(payload)
}

export async function unrecommendTeacherCommunityWriteup(
  id: string
): Promise<SubmissionWriteupData> {
  const payload = await request<
    SubmissionWriteupData & {
      id: string | number
      user_id: string | number
      challenge_id: string | number
      contest_id?: string | number
      recommended_by?: string | number
    }
  >({
    method: 'DELETE',
    url: `/teacher/community-writeups/${encodeURIComponent(id)}/recommend`,
  })
  return normalizeSubmissionWriteupData(payload)
}

export async function hideTeacherCommunityWriteup(id: string): Promise<SubmissionWriteupData> {
  const payload = await request<
    SubmissionWriteupData & {
      id: string | number
      user_id: string | number
      challenge_id: string | number
      contest_id?: string | number
      recommended_by?: string | number
    }
  >({
    method: 'POST',
    url: `/teacher/community-writeups/${encodeURIComponent(id)}/hide`,
  })
  return normalizeSubmissionWriteupData(payload)
}

export async function restoreTeacherCommunityWriteup(id: string): Promise<SubmissionWriteupData> {
  const payload = await request<
    SubmissionWriteupData & {
      id: string | number
      user_id: string | number
      challenge_id: string | number
      contest_id?: string | number
      recommended_by?: string | number
    }
  >({
    method: 'POST',
    url: `/teacher/community-writeups/${encodeURIComponent(id)}/restore`,
  })
  return normalizeSubmissionWriteupData(payload)
}

export async function getTeacherManualReviewSubmissions(params?: {
  student_id?: string
  challenge_id?: string
  class_name?: string
  review_status?: 'pending' | 'approved' | 'rejected'
  page?: number
  page_size?: number
}): Promise<PageResult<TeacherManualReviewSubmissionItemData>> {
  const payload = await request<PageResult<RawTeacherManualReviewSubmissionItem>>({
    method: 'GET',
    url: '/teacher/manual-review-submissions',
    params: {
      student_id: params?.student_id,
      challenge_id: params?.challenge_id,
      class_name: params?.class_name,
      review_status: params?.review_status,
      page: params?.page,
      page_size: params?.page_size,
    },
  })

  return {
    ...payload,
    list: payload.list.map((item) => ({
      ...item,
      id: String(item.id),
      user_id: String(item.user_id),
      challenge_id: String(item.challenge_id),
    })),
  }
}

export async function getTeacherManualReviewSubmission(
  id: string
): Promise<TeacherManualReviewSubmissionDetailData> {
  const payload = await request<RawTeacherManualReviewSubmissionDetail>({
    method: 'GET',
    url: `/teacher/manual-review-submissions/${encodeURIComponent(id)}`,
  })

  return {
    ...payload,
    id: String(payload.id),
    user_id: String(payload.user_id),
    challenge_id: String(payload.challenge_id),
  }
}

export async function reviewTeacherManualReviewSubmission(
  id: string,
  payload: {
    review_status: 'approved' | 'rejected'
    review_comment?: string
  }
): Promise<TeacherManualReviewSubmissionDetailData> {
  const response = await request<RawTeacherManualReviewSubmissionDetail>({
    method: 'PUT',
    url: `/teacher/manual-review-submissions/${encodeURIComponent(id)}/review`,
    data: payload,
  })

  return {
    ...response,
    id: String(response.id),
    user_id: String(response.user_id),
    challenge_id: String(response.challenge_id),
  }
}
