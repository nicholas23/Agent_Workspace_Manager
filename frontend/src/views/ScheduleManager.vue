<template>
  <div class="schedule-manager">
    <h2>全域排程管理</h2>
    <el-card>
      <div slot="header" class="clearfix">
        <span>等待中的任務</span>
        <el-button style="float: right; padding: 3px 0" type="text" @click="fetchSchedules">重新整理</el-button>
      </div>
      <el-table :data="schedules" style="width: 100%" v-loading="loading">
        <el-table-column prop="Project.name" label="專案名稱" width="180" />
        <el-table-column prop="command" label="指令" show-overflow-tooltip />
        <el-table-column prop="scheduled_time" label="預定執行時間" width="180">
          <template #default="scope">
            {{ new Date(scope.row.scheduled_time).toLocaleString() }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="100">
          <template #default="scope">
            <el-popconfirm title="確定要取消這個排程嗎？" @confirm="cancelSchedule(scope.row.id)">
              <template #reference>
                <el-button size="small" type="danger">取消</el-button>
              </template>
            </el-popconfirm>
          </template>
        </el-table-column>
      </el-table>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { ElMessage } from 'element-plus'

const schedules = ref([])
const loading = ref(false)

const fetchSchedules = async () => {
  loading.value = true
  try {
    const response = await axios.get('/api/schedules')
    schedules.value = response.data
  } catch (error) {
    console.error('Failed to fetch schedules:', error)
    ElMessage.error('無法載入排程列表')
  } finally {
    loading.value = false
  }
}

const cancelSchedule = async (id) => {
    // TODO: Implement cancel endpoint in backend if needed.
    // Currently Spec doesn't explicitly require cancel from global view, but it's good UX.
    // For now, just show a message that it's not implemented or implement the endpoint.
    // To be safe and stick to current scope, I'll just refresh.
    ElMessage.info('取消功能尚未實作')
}

onMounted(() => {
  fetchSchedules()
})
</script>

<style scoped>
.schedule-manager {
  padding: 20px;
}
</style>
