<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/api'

const route = useRoute()
const router = useRouter()
const id = computed(() => Number(route.params.id))

const runs = ref<any[]>([])
const indicators = ref<any[]>([])
const datapoints = ref<any[]>([])
const selected = ref<number | null>(null)

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

onMounted(reload)
</script>

<template>
  <div>
    <div class="bar">
      <h2>Collector #{{ id }} — Runs</h2>
      <el-button @click="router.push({ name: 'collector-edit', params: { id } })">Back to Edit</el-button>
    </div>

    <el-tabs>
      <el-tab-pane label="Runs">
        <el-table :data="runs" border @row-click="(r: any) => router.push({ name: 'run-detail', params: { id: r.id } })">
          <el-table-column prop="id" label="ID" width="80" />
          <el-table-column label="Status" width="120">
            <template #default="{ row }">
              <el-tag :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'info'">
                {{ row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="trigger" label="Trigger" width="120" />
          <el-table-column label="Started">
            <template #default="{ row }">{{ new Date(row.started_at).toLocaleString() }}</template>
          </el-table-column>
          <el-table-column label="Duration" width="120">
            <template #default="{ row }">{{ row.duration_ms }} ms</template>
          </el-table-column>
          <el-table-column prop="error" label="Error" />
        </el-table>
      </el-tab-pane>

      <el-tab-pane label="Data Points">
        <el-form inline>
          <el-form-item label="Indicator">
            <el-select v-model="selected" @change="loadDP" style="width: 240px">
              <el-option v-for="i in indicators" :key="i.id" :label="`${i.name} (${i.key})`" :value="i.id" />
            </el-select>
          </el-form-item>
        </el-form>
        <el-table :data="datapoints" border>
          <el-table-column label="Time">
            <template #default="{ row }">{{ new Date(row.ts).toLocaleString() }}</template>
          </el-table-column>
          <el-table-column label="Value">
            <template #default="{ row }">
              <span v-if="row.value_num !== null">{{ row.value_num }}</span>
              <span v-else-if="row.value_str !== null">{{ row.value_str }}</span>
              <span v-else>{{ row.value_json }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="run_id" label="Run" width="100" />
        </el-table>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<style scoped>
.bar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
</style>
