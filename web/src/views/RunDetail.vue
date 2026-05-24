<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api } from '@/api'

const route = useRoute()
const router = useRouter()
const id = computed(() => Number(route.params.id))

const run = ref<any>(null)
const logs = ref<any[]>([])

async function load() {
  const res = await api.runs.get(id.value)
  run.value = res.run
  logs.value = res.step_logs || []
}

onMounted(load)
</script>

<template>
  <div v-if="run">
    <div class="bar">
      <h2>Run #{{ run.id }}</h2>
      <el-button @click="router.back()">Back</el-button>
    </div>

    <el-descriptions border>
      <el-descriptions-item label="Collector">{{ run.collector_id }}</el-descriptions-item>
      <el-descriptions-item label="Status">
        <el-tag :type="run.status === 'success' ? 'success' : run.status === 'failed' ? 'danger' : 'info'">
          {{ run.status }}
        </el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="Trigger">{{ run.trigger }}</el-descriptions-item>
      <el-descriptions-item label="Started">{{ new Date(run.started_at).toLocaleString() }}</el-descriptions-item>
      <el-descriptions-item label="Duration">{{ run.duration_ms }} ms</el-descriptions-item>
      <el-descriptions-item v-if="run.error" label="Error">{{ run.error }}</el-descriptions-item>
    </el-descriptions>

    <h3 style="margin-top: 24px">Steps</h3>
    <el-timeline>
      <el-timeline-item v-for="s in logs" :key="s.id" :type="s.error ? 'danger' : 'success'" :timestamp="`#${s.index} · ${s.kind} · ${s.duration_ms} ms`">
        <div v-if="s.error" class="err">{{ s.error }}</div>
        <pre v-else class="snippet">{{ s.snippet || '(empty)' }}</pre>
      </el-timeline-item>
    </el-timeline>
  </div>
</template>

<style scoped>
.bar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.snippet { background: #f7f9fc; padding: 8px; border-radius: 4px; max-height: 240px; overflow: auto; font-size: 12px; }
.err { color: #c0392b; font-weight: 500; }
</style>
