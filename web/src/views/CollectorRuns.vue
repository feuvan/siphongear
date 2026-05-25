<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/api'
import ResponsiveTable, { type RTColumn } from '@/components/ResponsiveTable.vue'

const route = useRoute()
const router = useRouter()
const id = computed(() => Number(route.params.id))

const runs = ref<any[]>([])
const indicators = ref<any[]>([])
const datapoints = ref<any[]>([])
const selected = ref<number | null>(null)

const runColumns: RTColumn[] = [
  { key: 'id', label: 'ID', width: 80, primary: true },
  { key: 'status', label: 'Status', slot: 'status', width: 120 },
  { key: 'trigger', label: 'Trigger', width: 120 },
  { key: 'started', label: 'Started', slot: 'started' },
  { key: 'duration', label: 'Duration', slot: 'duration', width: 120 },
  { key: 'error', label: 'Error' }
]

const dpColumns: RTColumn[] = [
  { key: 'time', label: 'Time', slot: 'time', primary: true },
  { key: 'value', label: 'Value', slot: 'value' },
  { key: 'run_id', label: 'Run', width: 100 }
]

async function reload() {
  [runs.value, indicators.value] = await Promise.all([
    api.collectors.runs(id.value, 100),
    api.indicators.list(id.value)
  ])
  if (indicators.value.length && !selected.value) {
    selected.value = indicators.value[0].id
    await loadDP()
  }
}

async function loadDP() {
  if (!selected.value) return
  datapoints.value = await api.collectors.datapoints(id.value, { indicator_id: selected.value, limit: 200 })
}

function openRun(row: any) {
  router.push({ name: 'run-detail', params: { id: row.id } })
}

onMounted(reload)
</script>

<template>
  <div>
    <div class="page-bar">
      <h2>Collector #{{ id }} — Runs</h2>
      <div class="page-bar-actions">
        <el-button @click="router.push({ name: 'collector-edit', params: { id } })">Back to Edit</el-button>
      </div>
    </div>

    <el-tabs>
      <el-tab-pane label="Runs">
        <ResponsiveTable
          :rows="runs"
          :columns="runColumns"
          row-key="id"
          row-clickable
          @row-click="openRun"
        >
          <template #status="{ row }">
            <el-tag :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'info'">
              {{ row.status }}
            </el-tag>
          </template>
          <template #started="{ row }">{{ new Date(row.started_at).toLocaleString() }}</template>
          <template #duration="{ row }">{{ row.duration_ms }} ms</template>
        </ResponsiveTable>
      </el-tab-pane>

      <el-tab-pane label="Data Points">
        <el-form inline>
          <el-form-item label="Indicator">
            <el-select v-model="selected" @change="loadDP" style="width: 240px">
              <el-option v-for="i in indicators" :key="i.id" :label="`${i.name} (${i.key})`" :value="i.id" />
            </el-select>
          </el-form-item>
        </el-form>
        <ResponsiveTable :rows="datapoints" :columns="dpColumns">
          <template #time="{ row }">{{ new Date(row.ts).toLocaleString() }}</template>
          <template #value="{ row }">
            <span v-if="row.value_num !== null">{{ row.value_num }}</span>
            <span v-else-if="row.value_str !== null">{{ row.value_str }}</span>
            <span v-else>{{ row.value_json }}</span>
          </template>
        </ResponsiveTable>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>
