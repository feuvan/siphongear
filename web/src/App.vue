<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/store/auth'
import { useUiStore } from '@/store/ui'
import SideNav from '@/components/SideNav.vue'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const ui = useUiStore()

auth.loadFromStorage()
ui.init()

const isLogin = computed(() => route.name === 'login')

watch(() => route.fullPath, () => {
  if (ui.isMobile) ui.closeAside()
})

function logout() {
  auth.logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <el-container v-if="!isLogin" class="app-shell">
    <el-aside
      v-if="!ui.isMobile"
      :width="ui.asideOpen ? 'var(--sg-aside-w)' : 'var(--sg-aside-w-collapsed)'"
      class="aside"
    >
      <SideNav :collapsed="!ui.asideOpen" />
    </el-aside>

    <el-drawer
      v-else
      v-model="ui.asideOpen"
      direction="ltr"
      :with-header="false"
      size="240px"
      class="mobile-drawer"
    >
      <SideNav @navigate="ui.closeAside()" />
    </el-drawer>

    <el-container class="main-container">
      <el-header class="header">
        <el-button
          link
          class="icon-btn"
          @click="ui.toggleAside()"
          :title="ui.isMobile ? 'Menu' : (ui.asideOpen ? 'Collapse' : 'Expand')"
        >
          <el-icon :size="20"><Fold v-if="ui.asideOpen && !ui.isMobile" /><Expand v-else-if="!ui.isMobile" /><Menu v-else /></el-icon>
        </el-button>

        <div class="spacer"></div>

        <el-button
          link
          class="icon-btn"
          @click="ui.toggleTheme()"
          :title="ui.theme === 'dark' ? 'Switch to light' : 'Switch to dark'"
        >
          <el-icon :size="18">
            <Moon v-if="ui.theme === 'dark'" />
            <Sunny v-else />
          </el-icon>
        </el-button>

        <el-dropdown>
          <span class="user">
            <el-icon class="user-icon"><UserFilled /></el-icon>
            <span class="user-name">{{ auth.user?.username || 'guest' }}</span>
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="router.push({ name: 'settings' })">Settings</el-dropdown-item>
              <el-dropdown-item divided @click="logout">Logout</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
  <router-view v-else />
</template>

<style>
html,
body,
#app {
  height: 100%;
  margin: 0;
}
.app-shell {
  height: 100vh;
}
.aside {
  background: var(--sg-aside-bg);
  transition: width 0.2s ease;
  overflow: hidden;
}
.main-container {
  background: var(--sg-bg-page);
}
.header {
  display: flex;
  align-items: center;
  gap: 8px;
  height: var(--sg-header-h) !important;
  padding: 0 16px !important;
  border-bottom: 1px solid var(--sg-border-soft);
  background: var(--sg-bg-card);
}
.spacer {
  flex: 1;
}
.icon-btn {
  width: 36px;
  height: 36px;
  border-radius: 8px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  color: var(--sg-text-secondary);
}
.icon-btn:hover {
  background: var(--sg-aside-hover-bg);
  color: var(--sg-text-primary);
}
.user {
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 10px;
  border-radius: 8px;
  color: var(--sg-text-primary);
  font-size: 14px;
}
.user:hover {
  background: var(--sg-aside-hover-bg);
}
.user-icon {
  color: var(--sg-text-secondary);
}
.mobile-drawer :deep(.el-drawer__body) {
  padding: 0;
}
@media (max-width: 768px) {
  .user-name {
    display: none;
  }
  .header {
    padding: 0 10px !important;
  }
}
</style>
