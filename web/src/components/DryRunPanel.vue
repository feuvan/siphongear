<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ result: any }>()

const stepLogs = computed(() => props.result?.step_logs || [])
const indicators = computed(() => props.result?.indicators || {})
const vars = computed(() => props.result?.vars || {})
const errorText = computed(() => props.result?.error || '')
</script>

<template>
  <div v-if="result">
    <el-alert v-if="errorText" type="error" :closable="false" show-icon :title="errorText" style="margin-bottom: 12px;" />
    <el-alert v-else type="success" :closable="false" show-icon title="Pipeline OK" style="margin-bottom: 12px;" />

    <el-divider>Indicators</el-divider>
    <pre class="json">{{ JSON.stringify(indicators, null, 2) }}</pre>

    <el-divider>Final Vars</el-divider>
    <pre class="json">{{ JSON.stringify(vars, null, 2) }}</pre>

    <el-divider>Steps</el-divider>
    <el-timeline>
      <el-timeline-item v-for="(s, i) in stepLogs" :key="i" :type="s.error ? 'danger' : 'success'" :timestamp="`#${s.index} · ${s.kind} · ${s.duration_ms} ms`">
        <div v-if="s.error" class="err">{{ s.error }}</div>
        <pre v-else class="snippet">{{ s.snippet || '(empty)' }}</pre>
      </el-timeline-item>
    </el-timeline>
  </div>
  <el-empty v-else description="Run dry run first" />
</template>

<style scoped>
.json, .snippet { background: #f7f9fc; padding: 8px; border-radius: 4px; max-height: 200px; overflow: auto; font-size: 12px; }
.err { color: #c0392b; font-weight: 500; }
</style>
