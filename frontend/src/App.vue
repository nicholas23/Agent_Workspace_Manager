<template>
  <el-container class="layout-container">
    <el-header>
      <el-menu mode="horizontal" router :default-active="$route.path">
        <el-menu-item index="/">
          <span class="logo">AI Agent Manager</span>
        </el-menu-item>
        <el-menu-item index="/">Projects</el-menu-item>
        <el-menu-item index="/schedules">Schedules</el-menu-item>

      </el-menu>
    </el-header>
    <el-main>
      <router-view />
    </el-main>
  </el-container>
</template>

<style scoped>
.layout-container {
  height: 100vh;
}
.logo {
  font-weight: bold;
  font-size: 1.2rem;
  margin-right: 20px;
}
</style>