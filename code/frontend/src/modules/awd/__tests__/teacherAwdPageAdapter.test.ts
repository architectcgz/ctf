import { describe, expect, it } from 'vitest'

import type { TeacherAWDReviewArchiveData } from '@/api/contracts'
import { buildTeacherAwdPageModel } from '@/modules/awd/adapters/teacherAwdPageAdapter'

const review: TeacherAWDReviewArchiveData = {
  generated_at: '2026-04-20T12:00:00Z',
  scope: {
    snapshot_type: 'final',
    requested_by: 5,
    requested_id: 'contest-5',
  },
  contest: {
    id: 'contest-5',
    title: '春季 AWD 教学复盘',
    mode: 'awd',
    status: 'ended',
    current_round: 6,
    round_count: 6,
    team_count: 8,
    latest_evidence_at: '2026-04-20T11:58:00Z',
    export_ready: true,
  },
  overview: {
    round_count: 6,
    team_count: 8,
    service_count: 24,
    attack_count: 40,
    traffic_count: 120,
    latest_evidence_at: '2026-04-20T11:58:00Z',
  },
  rounds: [
    {
      id: 'round-6',
      contest_id: 'contest-5',
      round_number: 6,
      status: 'finished',
      started_at: '2026-04-20T11:00:00Z',
      ended_at: '2026-04-20T11:05:00Z',
      attack_score: 60,
      defense_score: 40,
      service_count: 24,
      attack_count: 12,
      traffic_count: 40,
    },
  ],
  selected_round: {
    round: {
      id: 'round-6',
      contest_id: 'contest-5',
      round_number: 6,
      status: 'finished',
      started_at: '2026-04-20T11:00:00Z',
      ended_at: '2026-04-20T11:05:00Z',
      attack_score: 60,
      defense_score: 40,
      service_count: 24,
      attack_count: 12,
      traffic_count: 40,
    },
    teams: [
      {
        team_id: 'team-1',
        team_name: 'Blue Fox',
        captain_id: 'captain-1',
        total_score: 420,
        member_count: 4,
        last_solve_at: '2026-04-20T11:04:00Z',
      },
    ],
    services: [
      {
        id: 'service-1',
        round_id: 'round-6',
        team_id: 'team-1',
        team_name: 'Blue Fox',
        service_id: 'runtime-1',
        challenge_id: 'challenge-1',
        challenge_title: 'gateway',
        service_status: 'up',
        attack_received: 1,
        sla_score: 30,
        defense_score: 20,
        attack_score: 10,
        updated_at: '2026-04-20T11:03:00Z',
      },
    ],
    attacks: [
      {
        id: 'attack-1',
        round_id: 'round-6',
        attacker_team_id: 'team-1',
        attacker_team_name: 'Blue Fox',
        victim_team_id: 'team-2',
        victim_team_name: 'Red Wolf',
        service_id: 'runtime-2',
        challenge_id: 'challenge-2',
        challenge_title: 'bank-portal',
        attack_type: 'flag_capture',
        source: 'submission',
        submitted_flag: 'flag{demo}',
        is_success: true,
        score_gained: 60,
        created_at: '2026-04-20T11:04:00Z',
      },
    ],
    traffic: [
      {
        id: 'traffic-1',
        contest_id: 'contest-5',
        round_id: 'round-6',
        attacker_team_id: 'team-1',
        attacker_team_name: 'Blue Fox',
        victim_team_id: 'team-2',
        victim_team_name: 'Red Wolf',
        service_id: 'runtime-2',
        challenge_id: 'challenge-2',
        challenge_title: 'bank-portal',
        method: 'POST',
        path: '/api/login',
        status_code: 500,
        source: 'nginx',
        created_at: '2026-04-20T11:04:10Z',
      },
    ],
  },
}

describe('buildTeacherAwdPageModel', () => {
  it('maps review archive data to overview, teams, services, replay, and export pages', () => {
    const result = buildTeacherAwdPageModel({
      review,
      selectedPage: 'services',
    })

    expect(result.hero.pageTitle).toBe('Service 复盘')
    expect(result.services.cards.length).toBeGreaterThan(0)
    expect(result.export.canExportReport).toBe(true)
    expect(result.replay.timeline[0].title).toContain('第 6 轮')
  })
})
