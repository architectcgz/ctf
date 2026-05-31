import { readFileSync } from 'node:fs'
import { join } from 'node:path'

export type FrontendArchitectureLayer =
  | 'common'
  | 'shared'
  | 'entities'
  | 'pages'
  | 'features'
  | 'widgets'

export interface FrontendArchitecturePolicy {
  layers: Record<
    FrontendArchitectureLayer,
    {
      relative_prefixes: string[]
      forbidden_import_layers: FrontendArchitectureLayer[]
    }
  >
  route_page_line_limit: number
  legacy_business_component_directories: string[]
  component_api_contract_only: boolean
  widget_api_contract_only: boolean
  feature_ui_api_contract_only: boolean
  low_level_forbidden_import_prefixes: string[]
  low_level_forbidden_bare_imports: string[]
  store_forbidden_import_prefixes: string[]
  utility_forbidden_import_prefixes: string[]
  utility_forbidden_bare_imports: string[]
  feature_router_forbidden_imports: string[]
  route_page: {
    allow_api_contract_imports_only: boolean
    forbidden_runtime_hooks: string[]
  }
  route_page_layer: FrontendArchitectureLayer
  route_page_suffix: string
  growth_baseline_file: string
  feature_boundary_guard_tests: string[]
}

const policyPath = join(process.cwd(), 'scripts', 'frontend-architecture-policy.json')

export const frontendArchitecturePolicy = JSON.parse(
  readFileSync(policyPath, 'utf8')
) as FrontendArchitecturePolicy

export const routePageLineLimit = frontendArchitecturePolicy.route_page_line_limit
