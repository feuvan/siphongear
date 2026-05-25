<script setup lang="ts">
import { computed } from 'vue'
import { useUiStore } from '@/store/ui'

export interface RTColumn {
  key: string
  label: string
  width?: string | number
  slot?: string
  hideOnMobile?: boolean
  primary?: boolean
}

const props = withDefaults(defineProps<{
  rows: any[]
  columns: RTColumn[]
  loading?: boolean
  rowKey?: string
  border?: boolean
  empty?: string
  rowClickable?: boolean
}>(), {
  border: true,
  empty: 'No data',
  rowClickable: false
})

const emit = defineEmits<{
  (e: 'row-click', row: any): void
}>()

const ui = useUiStore()

const visibleMobileCols = computed(() => props.columns.filter(c => !c.hideOnMobile))
const primaryCol = computed(() => props.columns.find(c => c.primary) || props.columns[0])
const otherCols = computed(() =>
  visibleMobileCols.value.filter(c => c.key !== primaryCol.value?.key && c.key !== 'actions')
)
const actionsCol = computed(() => props.columns.find(c => c.key === 'actions'))

function getCell(row: any, col: RTColumn) {
  return row?.[col.key]
}
</script>

<template>
  <div>
    <el-table
      v-if="!ui.isMobile"
      :data="rows"
      v-loading="loading"
      :border="border"
      style="width: 100%"
      @row-click="(r: any) => emit('row-click', r)"
    >
      <el-table-column
        v-for="col in columns"
        :key="col.key"
        :prop="col.slot ? undefined : col.key"
        :label="col.label"
        :width="col.width"
      >
        <template v-if="col.slot" #default="{ row }">
          <slot :name="col.slot" :row="row" />
        </template>
      </el-table-column>
    </el-table>

    <div v-else class="rt-mobile">
      <el-empty v-if="!rows.length && !loading" :description="empty" />
      <div
        v-for="(row, idx) in rows"
        :key="rowKey ? row[rowKey] : idx"
        class="rt-card"
        :class="{ clickable: rowClickable }"
        @click="rowClickable && emit('row-click', row)"
      >
        <div v-if="primaryCol" class="rt-card-title">
          <slot v-if="primaryCol.slot" :name="primaryCol.slot" :row="row" />
          <template v-else>{{ getCell(row, primaryCol) }}</template>
        </div>
        <div class="rt-card-body">
          <div v-for="col in otherCols" :key="col.key" class="rt-row">
            <span class="rt-label">{{ col.label }}</span>
            <span class="rt-value">
              <slot v-if="col.slot" :name="col.slot" :row="row" />
              <template v-else>{{ getCell(row, col) ?? '—' }}</template>
            </span>
          </div>
        </div>
        <div v-if="actionsCol" class="rt-actions" @click.stop>
          <slot :name="actionsCol.slot || 'actions'" :row="row" />
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.rt-mobile {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.rt-card {
  background: var(--sg-bg-card);
  border: 1px solid var(--sg-border-soft);
  border-radius: var(--sg-radius);
  padding: 12px 14px;
  box-shadow: var(--ep-box-shadow-light);
}
.rt-card.clickable {
  cursor: pointer;
}
.rt-card.clickable:active {
  background: var(--sg-aside-hover-bg);
}
.rt-card-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--sg-text-primary);
  margin-bottom: 8px;
  word-break: break-all;
}
.rt-card-body {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 8px;
}
.rt-row {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 10px;
  font-size: 13px;
}
.rt-label {
  color: var(--sg-text-secondary);
  flex-shrink: 0;
}
.rt-value {
  color: var(--sg-text-primary);
  text-align: right;
  word-break: break-all;
}
.rt-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding-top: 8px;
  border-top: 1px dashed var(--sg-border-soft);
}
.rt-actions :deep(.el-button) {
  padding: 6px 8px;
}
</style>
