import { describe, expect, it } from 'vitest'

import type {
  AWDAttackLogData,
  AWDReadinessData,
  AWDRoundData,
  AWDRoundSummaryData,
  AWDTrafficEventData,
  AWDTrafficSummaryData,
  AWDTeamServiceData,
  AdminContestChallengeData,
  AdminContestTeamData,
  ContestDetailData,
  ScoreboardRow,
} from '@/api/contracts'
import { buildAdminAwdPageModel } from '@/modules/awd/adapters/adminAwdPageAdapter'

const contest: ContestDetailData = {
  id: '9',
  title: '2026 全国 AWD 总决赛',
  description: '总决赛攻防场景',
  mode: 'awd',
  status: 'running',
  starts_at: '2026-04-20T08:00:00Z',
  ends_at: '2026-04-20T20:00:00Z',
  scoreboard_frozen: false,
}

const rounds: AWDRoundData[] = [
  {
    id: 'round-9',
    contest_id: '9',
    round_number: 9,
    status: 'running',
    attack_score: 60,
    defense_score: 40,
    created_at: '2026-04-20T10:00:00Z',
    updated_at: '2026-04-20T10:05:00Z',
  },
]

const summary: AWDRoundSummaryData = {
  round: rounds[0],
  metrics: {
    total_service_count: 12,
    service_up_count: 8,
    service_down_count: 2,
    service_compromised_count: 2,
    attacked_service_count: 5,
    defense_success_count: 8,
    total_attack_count: 22,
    successful_attack_count: 6,
    failed_attack_count: 16,
    scheduler_check_count: 12,
    manual_current_round_check_count: 1,
    manual_selected_round_check_count: 0,
    manual_service_check_count: 2,
    submission_attack_count: 12,
    manual_attack_log_count: 2,
    legacy_attack_log_count: 0,
  },
  items: [
    {
      team_id: 'team-1',
      team_name: 'Blue Fox',
      service_up_count: 4,
      service_down_count: 0,
      service_compromised_count: 0,
      sla_score: 90,
      defense_score: 80,
      attack_score: 60,
      successful_attack_count: 2,
      successful_breach_count: 0,
      unique_attackers_against: 1,
      total_score: 230,
    },
  ],
}

const readiness: AWDReadinessData = {
  contest_id: '9',
  ready: false,
  total_challenges: 4,
  passed_challenges: 2,
  pending_challenges: 1,
  failed_challenges: 1,
  stale_challenges: 0,
  missing_checker_challenges: 0,
  blocking_count: 1,
  global_blocking_reasons: [],
  blocking_actions: ['create_round'],
  items: [
    {
      challenge_id: 'challenge-2',
      title: 'bank-portal',
      checker_type: 'http_standard',
      validation_state: 'failed',
      blocking_reason: 'last_preview_failed',
      last_access_url: 'http://10.0.0.2',
      last_preview_at: '2026-04-20T09:58:00Z',
    },
  ],
}

const services: AWDTeamServiceData[] = [
  {
    id: 'service-row-1',
    round_id: 'round-9',
    team_id: 'team-1',
    team_name: 'Blue Fox',
    service_id: 'service-1',
    challenge_id: 'challenge-1',
    service_status: 'up',
    check_result: {},
    attack_received: 1,
    sla_score: 30,
    defense_score: 20,
    attack_score: 15,
    updated_at: '2026-04-20T10:04:00Z',
  },
  {
    id: 'service-row-2',
    round_id: 'round-9',
    team_id: 'team-2',
    team_name: 'Red Wolf',
    service_id: 'service-2',
    challenge_id: 'challenge-2',
    service_status: 'compromised',
    check_result: {},
    attack_received: 4,
    sla_score: 10,
    defense_score: 5,
    attack_score: 0,
    updated_at: '2026-04-20T10:04:30Z',
  },
]

const attacks: AWDAttackLogData[] = [
  {
    id: 'attack-1',
    round_id: 'round-9',
    attacker_team_id: 'team-1',
    attacker_team: 'Blue Fox',
    victim_team_id: 'team-2',
    victim_team: 'Red Wolf',
    service_id: 'service-2',
    challenge_id: 'challenge-2',
    attack_type: 'flag_capture',
    source: 'submission',
    submitted_flag: 'flag{demo}',
    is_success: true,
    score_gained: 60,
    created_at: '2026-04-20T10:03:00Z',
  },
]

const trafficSummary: AWDTrafficSummaryData = {
  contest_id: '9',
  round_id: 'round-9',
  round: rounds[0],
  total_request_count: 1200,
  active_attacker_team_count: 6,
  victim_team_count: 6,
  unique_path_count: 45,
  error_request_count: 120,
  latest_event_at: '2026-04-20T10:05:00Z',
  top_attackers: [
    {
      team_id: 'team-1',
      team_name: 'Blue Fox',
      request_count: 180,
      error_count: 12,
    },
  ],
  top_victims: [
    {
      team_id: 'team-2',
      team_name: 'Red Wolf',
      request_count: 210,
      error_count: 30,
    },
  ],
  top_challenges: [
    {
      challenge_id: 'challenge-2',
      challenge_title: 'bank-portal',
      request_count: 500,
      error_count: 90,
    },
  ],
  top_paths: [
    {
      path: '/api/login',
      request_count: 260,
      error_count: 20,
      last_status_code: 500,
    },
  ],
  top_error_paths: [
    {
      path: '/api/login',
      request_count: 260,
      error_count: 20,
      last_status_code: 500,
    },
  ],
  trend_buckets: [
    {
      bucket_start_at: '2026-04-20T10:00:00Z',
      bucket_end_at: '2026-04-20T10:05:00Z',
      request_count: 1200,
      error_count: 120,
    },
  ],
}

const trafficEvents: AWDTrafficEventData[] = [
  {
    id: 'traffic-1',
    contest_id: '9',
    round_id: 'round-9',
    attacker_team_id: 'team-1',
    attacker_team_name: 'Blue Fox',
    victim_team_id: 'team-2',
    victim_team_name: 'Red Wolf',
    service_id: 'service-2',
    challenge_id: 'challenge-2',
    challenge_title: 'bank-portal',
    method: 'POST',
    path: '/api/login',
    status_code: 500,
    status_group: 'server_error',
    is_error: true,
    source: 'nginx',
    occurred_at: '2026-04-20T10:04:40Z',
  },
]

const teams: AdminContestTeamData[] = [
  {
    id: 'team-1',
    contest_id: '9',
    name: 'Blue Fox',
    captain_id: 'captain-1',
    invite_code: 'AAA111',
    max_members: 5,
    member_count: 4,
    created_at: '2026-04-20T08:10:00Z',
  },
  {
    id: 'team-2',
    contest_id: '9',
    name: 'Red Wolf',
    captain_id: 'captain-2',
    invite_code: 'BBB222',
    max_members: 5,
    member_count: 5,
    created_at: '2026-04-20T08:12:00Z',
  },
]

const challengeLinks: AdminContestChallengeData[] = [
  {
    id: 'link-1',
    contest_id: '9',
    challenge_id: 'challenge-1',
    title: 'gateway',
    category: 'web',
    difficulty: 'medium',
    points: 100,
    order: 1,
    is_visible: true,
    awd_service_id: 'service-1',
    awd_service_display_name: 'gateway',
    awd_checker_type: 'http_standard',
    awd_sla_score: 30,
    awd_defense_score: 20,
    awd_checker_validation_state: 'passed',
    created_at: '2026-04-20T08:20:00Z',
  },
  {
    id: 'link-2',
    contest_id: '9',
    challenge_id: 'challenge-2',
    title: 'bank-portal',
    category: 'web',
    difficulty: 'hard',
    points: 200,
    order: 2,
    is_visible: true,
    awd_service_id: 'service-2',
    awd_service_display_name: 'bank-portal',
    awd_checker_type: 'http_standard',
    awd_sla_score: 40,
    awd_defense_score: 20,
    awd_checker_validation_state: 'failed',
    created_at: '2026-04-20T08:22:00Z',
  },
]

const scoreboardRows: ScoreboardRow[] = [
  {
    rank: 1,
    team_id: 'team-1',
    team_name: 'Blue Fox',
    score: 230,
    solved_count: 0,
    last_submission_at: '2026-04-20T10:03:00Z',
  },
]

describe('buildAdminAwdPageModel', () => {
  it('builds page models from readiness, round summary, traffic, attacks, and services', () => {
    const result = buildAdminAwdPageModel({
      contest,
      rounds,
      summary,
      readiness,
      services,
      attacks,
      trafficSummary,
      trafficEvents,
      teams,
      challengeLinks,
      scoreboardRows,
      selectedPage: 'alerts',
    })

    expect(result.hero.pageTitle).toBe('告警中心')
    expect(result.alerts.items.length).toBeGreaterThan(0)
    expect(result.instances.rows.length).toBeGreaterThan(0)
    expect(result.replay.timeline[0].title).toContain('第 9 轮')
  })
})
