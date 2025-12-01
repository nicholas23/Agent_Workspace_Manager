<template>
  <div>
    <div class="header-actions">
      <h2>專案列表</h2>
      <el-button type="primary" @click="dialogVisible = true">建立專案</el-button>
    </div>

    <!-- 專案列表表格 -->
    <el-table :data="store.projects" style="width: 100%" v-loading="store.loading">
      <el-table-column prop="name" label="名稱" width="180" />
      <el-table-column prop="ai_cli_command" label="AI 指令" width="200" />
      <el-table-column label="描述" width="150">
        <template #default="scope">
          {{ scope.row.description ? (scope.row.description.length > 10 ? scope.row.description.substring(0, 10) + '...' : scope.row.description) : '' }}
        </template>
      </el-table-column>
      <el-table-column prop="directory_path" label="路徑" />
      <el-table-column label="操作" width="200">
        <template #default="scope">
          <el-button size="small" @click="$router.push(`/projects/${scope.row.ID}`)">查看</el-button>
          <el-button size="small" type="danger" @click="handleDelete(scope.row.ID)">刪除</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 建立專案對話框 -->
    <el-dialog v-model="dialogVisible" title="建立專案">
      <el-form :model="form" label-width="120px">
        <el-form-item label="名稱">
          <el-input v-model="form.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" />
        </el-form-item>
        <el-form-item label="AI 指令">
          <el-input v-model="form.ai_cli_command" placeholder="例如：gemini 或 aider" />
        </el-form-item>
        <el-form-item label="目錄路徑">
          <el-input v-model="form.directory_path" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleCreate">建立</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, reactive } from 'vue'
import { useProjectStore } from '../stores/project'
import { ElMessage, ElMessageBox } from 'element-plus'

const store = useProjectStore()
const dialogVisible = ref(false)
const form = reactive({
  name: '',
  description: '',
  ai_cli_command: '',
  directory_path: ''
})

// 元件掛載時取得專案列表
onMounted(() => {
  store.fetchProjects()
})

// 處理建立專案
const handleCreate = async () => {
  try {
    await store.createProject(form)
    dialogVisible.value = false
    ElMessage.success('專案建立成功')
    // 重置表單
    form.name = ''
    form.description = ''
    form.ai_cli_command = ''
    form.directory_path = ''
  } catch (err) {
    ElMessage.error(store.error)
  }
}

// 處理刪除專案
const handleDelete = (id) => {
  ElMessageBox.confirm('確定要刪除此專案嗎？', '警告', {
    confirmButtonText: '確定',
    cancelButtonText: '取消',
    type: 'warning',
  }).then(async () => {
    await store.deleteProject(id)
    ElMessage.success('刪除成功')
  })
}
</script>

<style scoped>
.header-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
</style>
