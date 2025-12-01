<template>
  <div>
    <div class="header-actions">
      <h2>執行歷史</h2>
      <div class="filters">
        <el-select v-model="filterStatus" placeholder="狀態篩選" clearable style="width: 150px; margin-right: 10px;">
          <el-option label="執行中" value="running" />
          <el-option label="已完成" value="completed" />
          <el-option label="失敗" value="failed" />
          <el-option label="解析失敗" value="parse_failed" />
        </el-select>
        <el-button @click="fetchExecutions">重新整理</el-button>
      </div>
    </div>

    <!-- 執行記錄表格 -->
    <el-table :data="filteredExecutions" style="width: 100%" v-loading="loading">
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="status" label="狀態" width="100">
        <template #default="scope">
          <el-tag :type="getStatusType(scope.row.status)">{{ scope.row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="command" label="指令" show-overflow-tooltip />
      <el-table-column prop="summary" label="摘要" show-overflow-tooltip />
      <el-table-column prop="start_time" label="開始時間" width="180">
        <template #default="scope">
          {{ new Date(scope.row.start_time).toLocaleString() }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100">
        <template #default="scope">
          <el-button size="small" @click="viewDetails(scope.row)">詳情</el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 執行詳情對話框 -->
    <el-dialog v-model="detailVisible" title="執行詳情" width="70%">
      <div v-if="currentExecution">
        <h3>指令</h3>
        <pre>{{ currentExecution.command }}</pre>
        <h3>輸出</h3>
        <pre class="output-log">{{ currentExecution.details }}</pre>
        <div v-if="currentExecution.error_message">
          <h3>錯誤</h3>
          <pre class="error-log">{{ currentExecution.error_message }}</pre>
        </div>
        <div v-if="currentExecution.modified_files && currentExecution.modified_files.length">
          <h3>修改檔案</h3>
          <ul>
            <li v-for="file in currentExecution.modified_files" :key="file">{{ file }}</li>
          </ul>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRoute } from 'vue-router'
import axios from 'axios'

const route = useRoute()
const executions = ref([])
const loading = ref(false)
const detailVisible = ref(false)
const currentExecution = ref(null)
const filterStatus = ref('')

// 取得執行記錄
const fetchExecutions = async () => {
  loading.value = true
  try {
    const response = await axios.get(`/api/projects/${route.params.id}/executions`)
    executions.value = response.data
  } catch (err) {
    console.error(err)
  } finally {
    loading.value = false
  }
}

// 計算過濾後的列表
const filteredExecutions = computed(() => {
  if (!filterStatus.value) return executions.value
  return executions.value.filter(e => e.status === filterStatus.value)
})

onMounted(() => {
  fetchExecutions()
})

// 查看詳情
const viewDetails = (row) => {
  currentExecution.value = row
  detailVisible.value = true
}

// 取得狀態對應的標籤類型
const getStatusType = (status) => {
  switch (status) {
    case 'completed': return 'success'
    case 'failed': return 'danger'
    case 'running': return 'warning'
    default: return 'info'
  }
}
</script>

<style scoped>
.header-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.filters {
  display: flex;
  align-items: center;
}
.output-log {
  background: #f5f7fa;
  padding: 10px;
  border-radius: 4px;
  max-height: 400px;
  overflow: auto;
}
.error-log {
  background: #fef0f0;
  color: #f56c6c;
  padding: 10px;
  border-radius: 4px;
}
</style>
