import { defineStore } from 'pinia'
import axios from 'axios'

// 定義 Project Store 用於管理專案相關狀態
export const useProjectStore = defineStore('project', {
    state: () => ({
        projects: [],       // 專案列表
        currentProject: null, // 當前選中的專案
        loading: false,     // 載入狀態
        error: null         // 錯誤訊息
    }),
    actions: {
        // 取得所有專案
        async fetchProjects() {
            this.loading = true
            try {
                const response = await axios.get('/api/projects')
                this.projects = response.data
            } catch (err) {
                this.error = err.message
            } finally {
                this.loading = false
            }
        },
        // 取得單一專案詳情
        async fetchProject(id) {
            this.loading = true
            try {
                const response = await axios.get(`/api/projects/${id}`)
                this.currentProject = response.data
            } catch (err) {
                this.error = err.message
            } finally {
                this.loading = false
            }
        },
        // 建立新專案
        async createProject(project) {
            this.loading = true
            try {
                await axios.post('/api/projects', project)
                await this.fetchProjects() // 重新整理列表
            } catch (err) {
                this.error = err.response?.data?.error || err.message
                throw err
            } finally {
                this.loading = false
            }
        },
        // 刪除專案
        async deleteProject(id) {
            try {
                await axios.delete(`/api/projects/${id}`)
                await this.fetchProjects() // 重新整理列表
            } catch (err) {
                this.error = err.message
            }
        },
        // 執行專案指令
        async runCommand(id, command) {
            try {
                await axios.post(`/api/projects/${id}/run`, { command })
            } catch (err) {
                this.error = err.response?.data?.error || err.message
                throw err
            }
        },
        // 取得專案排程
        async fetchSchedules(id) {
            try {
                const response = await axios.get(`/api/projects/${id}/schedules`)
                return response.data
            } catch (err) {
                this.error = err.message
                return []
            }
        },
        // 建立專案排程
        async createSchedule(id, schedule) {
            try {
                await axios.post(`/api/projects/${id}/schedules`, schedule)
            } catch (err) {
                this.error = err.response?.data?.error || err.message
                throw err
            }
        }
    }
})
