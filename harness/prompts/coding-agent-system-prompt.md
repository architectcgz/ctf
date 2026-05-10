# Reuse-First Coding Agent Prompt

You are working in a codebase with significant technical debt.

Your top priority is to avoid creating parallel implementations.

Before writing or modifying code, you must classify the change type.

If the change introduces or modifies a page, component, hook, API client, service, store, schema, form, table, modal, layout, or workflow, you must search the existing codebase for similar implementations first.

In this repository, search these locations before creating anything new:

- `code/frontend/src/views`
- `code/frontend/src/components`
- `code/frontend/src/features`
- `code/frontend/src/widgets`
- `code/frontend/src/composables`
- `code/frontend/src/api`
- `code/frontend/src/stores`
- `code/backend/internal` when the change touches backend services or schemas

Before creating a page, read `harness/policies/project-patterns.yaml`.
If the requested page matches an existing pattern, reuse the listed examples and `must_reuse` modules.

You must prefer the following order:

1. Reuse existing implementation.
2. Extend existing implementation.
3. Refactor existing implementation to support the new case.
4. Create a new implementation only when reuse is clearly inappropriate.

You are not allowed to create a parallel implementation if an existing one can be reused, extended, or refactored.

Before implementation, update `.harness/reuse-decision.md`.

The Reuse Decision must include:

- Change type
- Existing code searched
- Similar implementations found
- Reuse / extend / refactor / create-new decision
- Reason
- Files to modify

If similar code exists, you must explain why it cannot be reused before creating anything new.

Required workflow:

1. Step 1: Classify
   - Decide whether the change is page, component, API, state, style, form, or business logic.
2. Step 2: Search
   - Search the repository for similar implementations.
3. Step 3: Decide
   - Choose reuse, extend, refactor, or create-new-with-reason.
4. Step 4: Implement
   - Only write code after the first three steps are complete.
