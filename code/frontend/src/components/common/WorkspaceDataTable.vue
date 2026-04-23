<script setup lang="ts">
export interface WorkspaceDataTableColumn {
  key: string
  label: string
  align?: 'left' | 'center' | 'right'
  widthClass?: string
  headerClass?: string
  cellClass?: string
}

const props = defineProps<{
  columns: WorkspaceDataTableColumn[]
  rows: unknown[]
  rowKey: string | ((row: unknown, index: number) => string | number)
  rowClass?: string
}>()

function getRowKey(row: unknown, index: number): string | number {
  if (typeof props.rowKey === 'function') {
    return props.rowKey(row, index)
  }

  if (!row || typeof row !== 'object') return index

  const value = (row as Record<string, unknown>)[props.rowKey]
  return typeof value === 'string' || typeof value === 'number' ? value : index
}

function getAlignClass(align?: WorkspaceDataTableColumn['align']): string {
  switch (align) {
    case 'center':
      return 'workspace-data-table__cell--center'
    case 'right':
      return 'workspace-data-table__cell--right'
    default:
      return 'workspace-data-table__cell--left'
  }
}

function getCellValue(row: unknown, key: string): unknown {
  if (!row || typeof row !== 'object') return undefined
  return (row as Record<string, unknown>)[key]
}
</script>

<template>
  <div class="workspace-data-table-shell">
    <table class="workspace-data-table">
      <thead class="workspace-data-table__head">
        <tr>
          <th
            v-for="column in columns"
            :key="column.key"
            :class="[
              'workspace-data-table__cell',
              'workspace-data-table__head-cell',
              getAlignClass(column.align),
              column.widthClass,
              column.headerClass,
            ]"
            scope="col"
          >
            {{ column.label }}
          </th>
        </tr>
      </thead>

      <tbody class="workspace-data-table__body">
        <tr
          v-for="(row, index) in rows"
          :key="getRowKey(row, index)"
          :class="['workspace-data-table__row', rowClass]"
        >
          <td
            v-for="column in columns"
            :key="column.key"
            :class="[
              'workspace-data-table__cell',
              'workspace-data-table__body-cell',
              getAlignClass(column.align),
              column.cellClass,
            ]"
          >
            <slot
              :name="`cell-${column.key}`"
              :row="row"
              :index="index"
              :column="column"
              :value="getCellValue(row, column.key)"
            >
              {{ getCellValue(row, column.key) }}
            </slot>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<style scoped>
.workspace-data-table-shell {
  --workspace-table-line: color-mix(in srgb, var(--color-border-default) 70%, transparent);
  --workspace-table-line-strong: color-mix(in srgb, var(--color-border-default) 84%, transparent);
  --workspace-table-text: color-mix(in srgb, var(--color-text-primary) 94%, transparent);
  --workspace-table-muted: color-mix(in srgb, var(--color-text-secondary) 88%, transparent);
  --workspace-table-faint: color-mix(in srgb, var(--color-text-muted) 84%, transparent);
  --workspace-table-row-hover: color-mix(in srgb, var(--color-border-default) 14%, transparent);
  width: 100%;
  overflow-x: auto;
}

.workspace-data-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}

.workspace-data-table__cell {
  vertical-align: middle;
}

.workspace-data-table__head-cell {
  padding: 0.75rem 0.5rem;
  border-bottom: 1px solid var(--workspace-table-line-strong);
  font-size: 0.6875rem;
  font-weight: 800;
  letter-spacing: 0.05em;
  text-transform: uppercase;
  color: var(--workspace-table-faint);
}

.workspace-data-table__body-cell {
  padding: 0.95rem 0.5rem;
}

.workspace-data-table__row {
  border-bottom: 1px solid var(--workspace-table-line);
  background: transparent;
  transition: background-color 0.2s ease;
}

.workspace-data-table__row:hover {
  background: var(--workspace-table-row-hover);
}

.workspace-data-table__cell--left {
  text-align: left;
}

.workspace-data-table__cell--center {
  text-align: center;
}

.workspace-data-table__cell--right {
  text-align: right;
}

:global([data-theme='light']) .workspace-data-table-shell {
  --workspace-table-line: color-mix(in srgb, var(--color-border-default) 82%, transparent);
  --workspace-table-line-strong: color-mix(in srgb, var(--color-border-default) 92%, transparent);
  --workspace-table-text: var(--color-text-primary);
  --workspace-table-muted: var(--color-text-secondary);
  --workspace-table-faint: var(--color-text-muted);
  --workspace-table-row-hover: color-mix(in srgb, var(--color-border-default) 18%, transparent);
}

:global([data-theme='dark']) .workspace-data-table-shell {
  --workspace-table-line: color-mix(in srgb, var(--color-border-default) 42%, transparent);
  --workspace-table-line-strong: color-mix(in srgb, var(--color-border-default) 56%, transparent);
  --workspace-table-text: color-mix(in srgb, var(--color-text-primary) 92%, transparent);
  --workspace-table-muted: color-mix(in srgb, var(--color-text-secondary) 86%, transparent);
  --workspace-table-faint: color-mix(in srgb, var(--color-text-muted) 80%, transparent);
  --workspace-table-row-hover: color-mix(in srgb, var(--color-border-default) 10%, transparent);
}
</style>
