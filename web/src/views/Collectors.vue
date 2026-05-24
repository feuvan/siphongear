<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/api'

const router = useRouter()
const rows = ref<any[]>([])
const sites = ref<any[]>([])
const loading = ref(false)

async function reload() {
  loading.value = true
  try {
    [rows.value, sites.value] = await Promise.all([api.collectors.list(), api.sites.list()])
  } finally { loading.value = false }
}

const siteName = computed(() => (id: number) => sites.value.find(s => s.id === id)?.name || `#${id}`)

async function toggle(row: any) {
  await api.collectors.update(row.id, { ...row, enabled: !row.enabled })
  await reload()
}

async function trigger(row: any) {
  try {
    const res = await api.collectors.run(row.id, {})
    ElMessage.success(`run ${res.run.id}: ${res.run.status}`)
    await reload()
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || 'run failed')
  }
}

async function remove(row: any) {
  await ElMessageBox.confirm(`Delete collector "${row.name}"?`, 'Confirm')
  await api.collectors.remove(row.id)
  await reload()
}

function gotoNew() { router.push({ name: 'collector-new' }) }
function gotoEdit(row: any) { router.push({ name: 'collector-edit', params: { id: row.id } }) }
function gotoRuns(row: any) { router.push({ name: 'collector-runs', params: { id: row.id } }) }

onMounted(reload)
</script>

<template>
  <div>
    <div class="bar">
      <h2>Collectors</h2>
      <el-button type="primary" @click="gotoNew">New Collector</el-button>
    </div>
    <el-table :data="rows" border v-loading="loading">
      <el-table-column prop="id" label="ID" width="70" />
      <el-table-column prop="name" label="Name" />
      <el-table-column label="Site" width="180">
        <template #default="{ row }">{{ siteName(row.site_id) }}</template>
      </el-table-column>
      <el-table-column label="Schedule" width="220">
        <template #default="{ row }">
          <el-tag size="small">{{ row.schedule_type || 'none' }}</el-tag>
          <span v-if="row.schedule_spec" style="margin-left:6px;">{{ row.schedule_spec }}</span>
        </template>
      </el-table-column>
      <el-table-column label="Status" width="140">
        <template #default="{ row }">
          <el-tag :type="row.last_status === 'success' ? 'success' : row.last_status === 'failed' ? 'danger' : 'info'">
            {{ row.last_status || '—' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="Enabled" width="110">
        <template #default="{ row }">
          <el-switch :model-value="row.enabled" @click="toggle(row)" />
        </template>
      </el-table-column>
      <el-table-column label="Actions" width="320">
        <template #default="{ row }">
          <el-button link type="primary" @click="trigger(row)">Run</el-button>
          <el-button link @click="gotoEdit(row)">Edit</el-button>
          <el-button link @click="gotoRuns(row)">Runs</el-button>
          <el-button link type="danger" @click="remove(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>
  </div>
</template>

<style scoped>
.bar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
</style>
