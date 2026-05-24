<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { api } from '@/api'

const rows = ref<any[]>([])
const sites = ref<any[]>([])
const dialog = ref(false)
const form = reactive<any>({ id: 0, site_id: 0, name: '', type: 'token', payload_text: '{}' })

async function reload() {
  [rows.value, sites.value] = await Promise.all([api.credentials.list(), api.sites.list()])
}

const siteName = computed(() => (id: number) => sites.value.find(s => s.id === id)?.name || `#${id}`)

function openCreate() {
  Object.assign(form, { id: 0, site_id: sites.value[0]?.id || 0, name: '', type: 'token', payload_text: '{}' })
  dialog.value = true
}

async function save() {
  let payload: any = {}
  try { payload = JSON.parse(form.payload_text || '{}') } catch (e: any) {
    ElMessage.error('payload must be valid JSON: ' + e.message)
    return
  }
  const body = { site_id: form.site_id, name: form.name, type: form.type, payload }
  if (form.id) await api.credentials.update(form.id, body)
  else await api.credentials.create(body)
  dialog.value = false
  await reload()
  ElMessage.success('saved')
}

async function openEdit(row: any) {
  Object.assign(form, { ...row, payload_text: '{}' })
  dialog.value = true
}

async function remove(row: any) {
  await ElMessageBox.confirm(`Delete ${row.name}?`, 'Confirm')
  await api.credentials.remove(row.id)
  await reload()
}

onMounted(reload)
</script>

<template>
  <div>
    <div class="bar">
      <h2>Credentials</h2>
      <el-button type="primary" @click="openCreate">New Credential</el-button>
    </div>
    <el-table :data="rows" border>
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="name" label="Name" />
      <el-table-column label="Site" width="180">
        <template #default="{ row }">{{ siteName(row.site_id) }}</template>
      </el-table-column>
      <el-table-column prop="type" label="Type" width="120" />
      <el-table-column label="Updated" width="200">
        <template #default="{ row }">{{ new Date(row.updated_at).toLocaleString() }}</template>
      </el-table-column>
      <el-table-column label="Actions" width="220">
        <template #default="{ row }">
          <el-button link @click="openEdit(row)">Edit / Replace</el-button>
          <el-button link type="danger" @click="remove(row)">Delete</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="dialog" :title="form.id ? 'Edit Credential' : 'New Credential'" width="640px">
      <el-form label-width="120px">
        <el-form-item label="Site">
          <el-select v-model="form.site_id">
            <el-option v-for="s in sites" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="Name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="Type">
          <el-select v-model="form.type">
            <el-option label="password" value="password" />
            <el-option label="cookie" value="cookie" />
            <el-option label="token" value="token" />
            <el-option label="custom" value="custom" />
          </el-select>
        </el-form-item>
        <el-form-item label="Payload (JSON)">
          <el-input v-model="form.payload_text" type="textarea" :rows="8" placeholder='{"api_key": "..."}' />
        </el-form-item>
        <el-alert v-if="form.id" type="info" show-icon :closable="false">
          For security, existing payload is not displayed. Submit replaces it.
        </el-alert>
      </el-form>
      <template #footer>
        <el-button @click="dialog = false">Cancel</el-button>
        <el-button type="primary" @click="save">Save</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.bar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
</style>
