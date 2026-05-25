<script setup lang="ts">
import { reactive } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/store/auth'

const router = useRouter()
const auth = useAuthStore()
const form = reactive({ username: '', password: '' })

async function submit() {
  try {
    await auth.login(form.username, form.password)
    router.push({ name: 'dashboard' })
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.error || 'login failed')
  }
}
</script>

<template>
  <div class="login">
    <div class="login-card">
      <div class="brand">
        <div class="logo">SG</div>
        <h1>SiphonGear</h1>
        <p class="tagline">Configurable collection &amp; metrics platform</p>
      </div>
      <el-form @submit.prevent="submit" label-position="top" class="form">
        <el-form-item label="Username">
          <el-input v-model="form.username" autofocus placeholder="admin" />
        </el-form-item>
        <el-form-item label="Password">
          <el-input v-model="form.password" type="password" show-password @keyup.enter="submit" />
        </el-form-item>
        <el-button type="primary" size="large" @click="submit" class="submit">Sign In</el-button>
      </el-form>
    </div>
  </div>
</template>

<style scoped>
.login {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 20px;
  background:
    radial-gradient(circle at 20% 20%, rgba(99, 102, 241, 0.18), transparent 55%),
    radial-gradient(circle at 80% 80%, rgba(139, 92, 246, 0.18), transparent 55%),
    linear-gradient(135deg, #eef2ff 0%, #f8fafc 100%);
}
:global(html.dark) .login {
  background:
    radial-gradient(circle at 20% 20%, rgba(99, 102, 241, 0.18), transparent 55%),
    radial-gradient(circle at 80% 80%, rgba(139, 92, 246, 0.18), transparent 55%),
    linear-gradient(135deg, #0b1020 0%, #131a2e 100%);
}
.login-card {
  width: min(400px, 92vw);
  background: var(--sg-bg-card);
  border: 1px solid var(--sg-border-soft);
  border-radius: 16px;
  padding: 36px 32px 32px;
  box-shadow: 0 20px 50px rgba(15, 23, 42, 0.12);
}
.brand {
  text-align: center;
  margin-bottom: 24px;
}
.logo {
  width: 52px;
  height: 52px;
  border-radius: 14px;
  margin: 0 auto 14px;
  background: linear-gradient(135deg, #6366f1, #8b5cf6);
  color: #fff;
  font-weight: 700;
  font-size: 18px;
  letter-spacing: 1px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 20px rgba(99, 102, 241, 0.35);
}
.brand h1 {
  margin: 0 0 6px;
  font-size: 22px;
  font-weight: 600;
  color: var(--sg-text-primary);
  letter-spacing: -0.3px;
}
.tagline {
  margin: 0;
  font-size: 13px;
  color: var(--sg-text-secondary);
}
.form :deep(.el-form-item__label) {
  font-weight: 500;
  color: var(--sg-text-secondary);
}
.submit {
  width: 100%;
  margin-top: 4px;
  height: 42px;
  font-weight: 500;
}
</style>
