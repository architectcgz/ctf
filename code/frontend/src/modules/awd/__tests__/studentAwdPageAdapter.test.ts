import { describe, expect, it } from 'vitest'

import type {
  AWDAttackLogData,
  ContestAWDWorkspaceData,
  ContestChallengeItem,
  ContestDetailData,
  ScoreboardRow,
} from '@/api/contracts'
import { buildStudentAwdPageModel } from '@/modules/awd/adapters/studentAwdPageAdapter'

const contest: ContestDetailData = {
  id: '42',
  title: '2026 春季 AWD',
  description: '春季攻防对抗赛',
  mode: 'awd',
  status: 'running',
  starts_at: '2026-04-20T12:00:00Z',
  ends_at: '2026-04-20T18:00:00Z',
  scoreboard_frozen: false,
}

const challenges: ContestChallengeItem[] = [
  {
    id: 'c-1',
    challenge_id: 'challenge-1',
    awd_service_id: 'service-1',
    title: 'web-gateway',
    category: 'web',
    difficulty: 'medium',
    points: 100,
    solved_count: 0,
    is_solved: false,
  },
  {
    id: 'c-2',
    challenge_id: 'challenge-2',
    awd_service_id: 'service-2',
    title: 'db-core',
    category: 'misc',
    difficulty: 'hard',
    points: 200,
    solved_count: 0,
    is_solved: false,
  },
]

const workspace: ContestAWDWorkspaceData = {
  contest_id: '42',
  current_round: {
    id: 'round-12',
    contest_id: '42',
    round_number: 12,
    status: 'running',
    attack_score: 50,
    defense_score: 30,
    started_at: '2026-04-20T14:00:00Z',
    ended_at: '2026-04-20T14:05:00Z',
    created_at: '2026-04-20T14:00:00Z',
    updated_at: '2026-04-20T14:00:00Z',
  },
  my_team: {
    team_id: 'team-self',
    team_name: 'Team Nebula',
  },
  services: [
    {
      service_id: 'service-1',
      challenge_id: 'challenge-1',
      access_url: 'http://10.0.0.2:8080',
      service_status: 'compromised',
      attack_received: 3,
      sla_score: 20,
      defense_score: 10,
      attack_score: 50,
      checker_type: 'http_standard',
      updated_at: '2026-04-20T14:02:00Z',
    },
    {
      service_id: 'service-2',
      challenge_id: 'challenge-2',
      access_url: 'tcp://10.0.0.2:3306',
      service_status: 'up',
      attack_received: 0,
      sla_score: 30,
      defense_score: 40,
      attack_score: 0,
      checker_type: 'http_standard',
      updated_at: '2026-04-20T14:03:00Z',
    },
  ],
  targets: [
    {
      team_id: 'team-a',
      team_name: 'TeamA',
      services: [
        { service_id: 'service-1', challenge_id: 'challenge-1', access_url: 'http://10.10.0.3:8080' },
        { service_id: 'service-2', challenge_id: 'challenge-2', access_url: 'tcp://10.10.0.3:3306' },
      ],
    },
  ],
  recent_events: [
    {
      id: 'event-1',
      direction: 'attack_out',
      service_id: 'service-1',
      challenge_id: 'challenge-1',
      peer_team_id: 'team-a',
      peer_team_name: 'TeamA',
      is_success: true,
      score_gained: 50,
      created_at: '2026-04-20T14:04:00Z',
    },
  ],
}

const scoreboardRows: ScoreboardRow[] = [
  {
    rank: 1,
    team_id: 'team-self',
    team_name: 'Team Nebula',
    score: 1337,
    solved_count: 0,
    last_submission_at: '2026-04-20T14:04:00Z',
  },
]

const submitResult: AWDAttackLogData = {
  id: 'attack-1',
  round_id: 'round-12',
  attacker_team_id: 'team-self',
  attacker_team: 'Team Nebula',
  victim_team_id: 'team-a',
  victim_team: 'TeamA',
  service_id: 'service-1',
  challenge_id: 'challenge-1',
  attack_type: 'flag_capture',
  source: 'submission',
  submitted_flag: 'flag{demo}',
  is_success: true,
  score_gained: 50,
  created_at: '2026-04-20T14:04:02Z',
}

describe('buildStudentAwdPageModel', () => {
  it('maps workspace and scoreboard data into the 5 student awd pages', () => {
    const result = buildStudentAwdPageModel({
      contest,
      challenges,
      workspace,
      scoreboardRows,
      selectedPage: 'targets',
      submitResult,
    })

    expect(result.hero.pageTitle).toBe('目标目录')
    expect(result.targets.rows[0].teamName).toBe('TeamA')
    expect(result.targets.rows[0].accessUrl).toBe('http://10.10.0.3:8080')
    expect(result.attacks.recent[0].challengeTitle).toBe('web-gateway')
    expect(result.overview.defenseAlerts[0].statusLabel).toBe('失陷')
    expect(result.collab.priorities[0]).toContain('TeamA')
  })
})
