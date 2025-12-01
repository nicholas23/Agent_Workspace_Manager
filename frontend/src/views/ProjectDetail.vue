<template>
  <div v-if="store.currentProject">
    <div class="header-actions">
      <h2>{{ store.currentProject.name }}</h2>
      <div>
        <el-button @click="editDialogVisible = true">編輯專案</el-button>
        <el-button @click="$router.push(`/projects/${store.currentProject.ID}/executions`)">歷史記錄</el-button>
        <el-button type="warning" @click="scheduleDialogVisible = true">排程任務</el-button>
        <el-button type="primary" @click="runDialogVisible = true">執行指令</el-button>
      </div>
    </div>

    <!-- 專案資訊 -->
    <el-descriptions border :column="1">
      <el-descriptions-item label="描述">{{ store.currentProject.description }}</el-descriptions-item>
      <el-descriptions-item label="路徑">{{ store.currentProject.directory_path }}</el-descriptions-item>
      <el-descriptions-item label="AI 指令">{{ store.currentProject.ai_cli_command }}</el-descriptions-item>
    </el-descriptions>

    <!-- 最後一次執行結果 -->
    <div class="last-execution-section">
      <h3>最後一次執行</h3>
      <el-card v-if="lastExecution" shadow="hover">
        <template #header>
          <div class="execution-header">
            <span>{{ lastExecution.command }}</span>
            <el-tag :type="getStatusType(lastExecution.status)">{{ getStatusText(lastExecution.status) }}</el-tag>
          </div>
        </template>
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="執行時間">
            {{ new Date(lastExecution.start_time).toLocaleString() }}
          </el-descriptions-item>
          <el-descriptions-item label="結束時間" v-if="lastExecution.end_time && lastExecution.status !== 'running'">
            {{ new Date(lastExecution.end_time).toLocaleString() }}
          </el-descriptions-item>
          <el-descriptions-item label="摘要" :span="2" v-if="lastExecution.summary">
            {{ lastExecution.summary }}
          </el-descriptions-item>
        </el-descriptions>
        <el-collapse v-if="lastExecution.details" class="execution-details-collapse">
          <el-collapse-item title="執行內容" name="details">
            <pre>{{ lastExecution.details }}</pre>
          </el-collapse-item>
        </el-collapse>
        <div v-if="lastExecution.error_message" class="error-message">
          <el-alert type="error" :closable="false">
            {{ lastExecution.error_message }}
          </el-alert>
        </div>
      </el-card>
      <el-empty v-else description="尚無執行記錄" />
    </div>

    <!-- 排程列表 -->
    <div class="schedules-section">
      <h3>即將執行的排程</h3>
      <el-table :data="schedules" style="width: 100%">
        <el-table-column prop="command" label="指令" />
        <el-table-column prop="scheduled_time" label="時間">
          <template #default="scope">
            {{ new Date(scope.row.scheduled_time).toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="狀態">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)">{{ getStatusText(scope.row.status) }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 執行指令對話框 -->
    <el-dialog v-model="runDialogVisible" title="執行指令">
      <el-form :model="runForm">
        <el-form-item label="指令">
          <el-input v-model="runForm.command" type="textarea" :rows="4" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="runDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleRun">執行</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 排程任務對話框 -->
    <el-dialog v-model="scheduleDialogVisible" title="排程任務">
      <el-form :model="scheduleForm">
        <el-form-item label="指令">
          <el-input v-model="scheduleForm.command" type="textarea" :rows="2" />
        </el-form-item>
        <el-form-item label="時間">
          <el-date-picker v-model="scheduleForm.scheduled_time" type="datetime" placeholder="選擇日期時間" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="scheduleDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleSchedule">排程</el-button>
        </span>
      </template>
    </el-dialog>

    <!-- 編輯專案對話框 -->
    <el-dialog v-model="editDialogVisible" title="編輯專案">
      <el-form :model="editForm" label-width="120px">
        <el-form-item label="名稱">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="editForm.description" />
        </el-form-item>
        <el-form-item label="AI 指令">
          <el-input v-model="editForm.ai_cli_command" />
        </el-form-item>
        <el-form-item label="目錄路徑">
          <el-input v-model="editForm.directory_path" />
        </el-form-item>
      </el-form>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="editDialogVisible = false">取消</el-button>
          <el-button type="primary" @click="handleUpdate">儲存</el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, reactive, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useProjectStore } from '../stores/project'
import { ElMessage } from 'element-plus'
import axios from 'axios'

const route = useRoute()
const store = useProjectStore()
const runDialogVisible = ref(false)
const scheduleDialogVisible = ref(false)
const editDialogVisible = ref(false)
const schedules = ref([])
const lastExecution = ref(null)
let updateInterval = null

const runForm = reactive({
  command: ''
})

const scheduleForm = reactive({
  command: '',
  scheduled_time: ''
})

const editForm = reactive({
  name: '',
  description: '',
  ai_cli_command: '',
  directory_path: ''
})

// 載入排程列表
const loadSchedules = async () => {
  schedules.value = await store.fetchSchedules(route.params.id)
}

// 載入最後一次執行記錄
const loadLastExecution = async () => {
  try {
    const response = await axios.get(`/api/projects/${route.params.id}/executions`)
    if (response.data && response.data.length > 0) {
      // 取第一筆（最新的）
      lastExecution.value = response.data[0]
    }
  } catch (err) {
    console.error('Failed to load last execution:', err)
  }
}

// 元件掛載時載入專案資訊、排程與最後一次執行
onMounted(async () => {
  await store.fetchProject(route.params.id)
  await loadSchedules()
  await loadLastExecution()
  
  // 啟動定時更新（每5秒）
  updateInterval = setInterval(async () => {
    await loadLastExecution()
    await loadSchedules()
  }, 5000)
})

// 元件卸載時清除定時器
onUnmounted(() => {
  if (updateInterval) {
    clearInterval(updateInterval)
  }
})

// 當專案資料載入後，同步到編輯表單
watch(() => store.currentProject, (newVal) => {
  if (newVal) {
    editForm.name = newVal.name
    editForm.description = newVal.description
    editForm.ai_cli_command = newVal.ai_cli_command
    editForm.directory_path = newVal.directory_path
  }
})

// 處理執行指令
const handleRun = async () => {
  try {
    await store.runCommand(route.params.id, runForm.command)
    runDialogVisible.value = false
    ElMessage.success('指令已開始執行')
    runForm.command = ''
  } catch (err) {
    ElMessage.error(store.error)
  }
}

// 處理建立排程
const handleSchedule = async () => {
  try {
    if (!scheduleForm.scheduled_time) {
      ElMessage.error('請選擇時間')
      return
    }
    await store.createSchedule(route.params.id, {
      command: scheduleForm.command,
      scheduled_time: scheduleForm.scheduled_time
    })
    scheduleDialogVisible.value = false
    ElMessage.success('排程建立成功')
    scheduleForm.command = ''
    scheduleForm.scheduled_time = ''
    await loadSchedules()
  } catch (err) {
    ElMessage.error(store.error)
  }
}

// 處理更新專案
const handleUpdate = async () => {
  try {
    await axios.put(`/api/projects/${route.params.id}`, editForm)
    await store.fetchProject(route.params.id) // 重新載入
    editDialogVisible.value = false
    ElMessage.success('專案更新成功')
  } catch (err) {
    ElMessage.error('更新失敗: ' + (err.response?.data?.error || err.message))
  }
}

// 狀態轉換輔助函數
const getStatusType = (status) => {
  const typeMap = {
    'pending': 'info',
    'running': 'warning',
    'completed': 'success',
    'failed': 'danger',
    'parse_failed': 'warning'
  }
  return typeMap[status] || 'info'
}

const getStatusText = (status) => {
  const textMap = {
    'pending': '等待中',
    'running': '執行中',
    'completed': '已完成',
    'failed': '失敗',
    'parse_failed': '解析失敗'
  }
  return textMap[status] || status
}
</script>

<style scoped>
.header-actions {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.last-execution-section {
  margin-top: 30px;
}

.execution-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.execution-details-collapse {
  margin-top: 15px;
}

.execution-details-collapse pre {
  background: #f5f7fa;
  padding: 15px;
  border-radius: 4px;
  overflow-x: auto;
  max-height: 300px;
  font-size: 12px;
  line-height: 1.5;
  margin: 0;
}

.error-message {
  margin-top: 15px;
}

.schedules-section {
  margin-top: 30px;
}
</style>
