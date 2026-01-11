<template>
  <div class="cron-picker">
    <!-- 模式切换 -->
    <div class="mode-tabs">
      <el-radio-group v-model="mode" size="small" @change="handleModeChange">
        <el-radio-button value="simple">简易模式</el-radio-button>
        <el-radio-button value="advanced">高级模式</el-radio-button>
      </el-radio-group>
    </div>

    <!-- 简易模式 -->
    <div v-if="mode === 'simple'" class="simple-mode">
      <!-- 预设选项 -->
      <div class="preset-section">
        <div class="section-title">常用预设</div>
        <div class="preset-grid">
          <div
            v-for="preset in presets"
            :key="preset.value"
            class="preset-item"
            :class="{ active: selectedPreset === preset.value }"
            @click="selectPreset(preset)"
          >
            <el-icon><component :is="preset.icon" /></el-icon>
            <span class="preset-label">{{ preset.label }}</span>
            <span class="preset-desc">{{ preset.desc }}</span>
          </div>
        </div>
      </div>

      <!-- 自定义时间 -->
      <div class="custom-section">
        <div class="section-title">自定义时间</div>
        
        <!-- 执行频率 -->
        <div class="custom-row">
          <span class="row-label">执行频率</span>
          <el-select v-model="customConfig.frequency" placeholder="选择频率" @change="updateFromCustom">
            <el-option label="每分钟" value="minute" />
            <el-option label="每小时" value="hour" />
            <el-option label="每天" value="day" />
            <el-option label="每周" value="week" />
            <el-option label="每月" value="month" />
            <el-option label="间隔执行" value="interval" />
          </el-select>
        </div>

        <!-- 间隔执行配置 -->
        <div v-if="customConfig.frequency === 'interval'" class="custom-row">
          <span class="row-label">执行间隔</span>
          <div class="interval-config">
            <span>每</span>
            <el-input-number
              v-model="customConfig.intervalValue"
              :min="1"
              :max="999"
              size="small"
              controls-position="right"
              @change="updateFromCustom"
            />
            <el-select v-model="customConfig.intervalUnit" size="small" @change="updateFromCustom">
              <el-option label="秒" value="s" />
              <el-option label="分钟" value="m" />
              <el-option label="小时" value="h" />
            </el-select>
            <span>执行一次</span>
          </div>
        </div>

        <!-- 每小时配置 -->
        <div v-if="customConfig.frequency === 'hour'" class="custom-row">
          <span class="row-label">分钟</span>
          <el-select v-model="customConfig.minute" placeholder="选择分钟" @change="updateFromCustom">
            <el-option v-for="m in 60" :key="m-1" :label="`第 ${m-1} 分钟`" :value="m-1" />
          </el-select>
        </div>

        <!-- 每天配置 -->
        <template v-if="customConfig.frequency === 'day'">
          <div class="custom-row">
            <span class="row-label">时间</span>
            <el-time-picker
              v-model="customConfig.time"
              format="HH:mm"
              value-format="HH:mm"
              placeholder="选择时间"
              @change="updateFromCustom"
            />
          </div>
        </template>

        <!-- 每周配置 -->
        <template v-if="customConfig.frequency === 'week'">
          <div class="custom-row">
            <span class="row-label">星期</span>
            <el-checkbox-group v-model="customConfig.weekdays" @change="updateFromCustom">
              <el-checkbox-button v-for="(day, index) in weekDays" :key="index" :value="index">
                {{ day }}
              </el-checkbox-button>
            </el-checkbox-group>
          </div>
          <div class="custom-row">
            <span class="row-label">时间</span>
            <el-time-picker
              v-model="customConfig.time"
              format="HH:mm"
              value-format="HH:mm"
              placeholder="选择时间"
              @change="updateFromCustom"
            />
          </div>
        </template>

        <!-- 每月配置 -->
        <template v-if="customConfig.frequency === 'month'">
          <div class="custom-row">
            <span class="row-label">日期</span>
            <el-select v-model="customConfig.dayOfMonth" placeholder="选择日期" @change="updateFromCustom">
              <el-option v-for="d in 31" :key="d" :label="`每月 ${d} 日`" :value="d" />
            </el-select>
          </div>
          <div class="custom-row">
            <span class="row-label">时间</span>
            <el-time-picker
              v-model="customConfig.time"
              format="HH:mm"
              value-format="HH:mm"
              placeholder="选择时间"
              @change="updateFromCustom"
            />
          </div>
        </template>
      </div>
    </div>

    <!-- 高级模式 -->
    <div v-else class="advanced-mode">
      <div class="cron-input-wrapper">
        <el-input
          v-model="cronExpression"
          placeholder="输入 cron 表达式，如: 0 0 * * * 或 @every 1h"
          @input="handleAdvancedInput"
        >
          <template #prefix>
            <el-icon><Clock /></el-icon>
          </template>
        </el-input>
      </div>
      
      <div class="cron-help">
        <div class="help-title">
          <el-icon><InfoFilled /></el-icon>
          <span>表达式格式说明</span>
        </div>
        <div class="help-content">
          <div class="help-row">
            <code>* * * * *</code>
            <span>分 时 日 月 周</span>
          </div>
          <div class="help-row">
            <code>@every 1h30m</code>
            <span>每 1 小时 30 分钟</span>
          </div>
          <div class="help-row">
            <code>@daily</code>
            <span>每天午夜</span>
          </div>
          <div class="help-row">
            <code>@hourly</code>
            <span>每小时</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 结果预览 -->
    <div class="result-preview">
      <div class="preview-label">生成的表达式</div>
      <div class="preview-expression">
        <code>{{ cronExpression || '请选择或输入' }}</code>
        <el-button v-if="cronExpression" type="primary" link size="small" @click="copyExpression">
          <el-icon><DocumentCopy /></el-icon>
          复制
        </el-button>
      </div>
      <div v-if="cronDescription" class="preview-description">
        <el-icon><Clock /></el-icon>
        <span>{{ cronDescription }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Clock, InfoFilled, DocumentCopy, Timer, Sunny, Calendar, AlarmClock } from '@element-plus/icons-vue'

const props = defineProps<{
  modelValue: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const mode = ref<'simple' | 'advanced'>('simple')
const cronExpression = ref(props.modelValue || '')
const selectedPreset = ref('')

const weekDays = ['日', '一', '二', '三', '四', '五', '六']

const customConfig = reactive({
  frequency: 'day' as 'minute' | 'hour' | 'day' | 'week' | 'month' | 'interval',
  minute: 0,
  time: '00:00',
  weekdays: [1] as number[],
  dayOfMonth: 1,
  intervalValue: 1,
  intervalUnit: 'h' as 's' | 'm' | 'h'
})

const presets = [
  { value: '* * * * *', label: '每分钟', desc: '每分钟执行', icon: Timer },
  { value: '0 * * * *', label: '每小时', desc: '每小时整点', icon: AlarmClock },
  { value: '0 0 * * *', label: '每天', desc: '每天 00:00', icon: Sunny },
  { value: '0 0 * * 1', label: '每周一', desc: '每周一 00:00', icon: Calendar },
  { value: '0 0 1 * *', label: '每月', desc: '每月 1 日 00:00', icon: Calendar },
  { value: '@every 30m', label: '每30分钟', desc: '每隔30分钟', icon: Timer },
  { value: '@every 1h', label: '每1小时', desc: '每隔1小时', icon: AlarmClock },
  { value: '@every 6h', label: '每6小时', desc: '每隔6小时', icon: AlarmClock },
]

// 解析 cron 表达式生成描述
const cronDescription = computed(() => {
  const expr = cronExpression.value.trim()
  if (!expr) return ''
  
  // @every 语法
  if (expr.startsWith('@every ')) {
    const interval = expr.slice(7)
    return `每隔 ${interval} 执行一次`
  }
  
  // 特殊预设
  const specials: Record<string, string> = {
    '@yearly': '每年执行一次',
    '@annually': '每年执行一次',
    '@monthly': '每月执行一次',
    '@weekly': '每周执行一次',
    '@daily': '每天执行一次',
    '@midnight': '每天午夜执行',
    '@hourly': '每小时执行一次'
  }
  if (specials[expr]) return specials[expr]
  
  // 标准 cron 格式
  const parts = expr.split(/\s+/)
  if (parts.length !== 5) return '表达式格式有误'
  
  const [minute, hour, day, month, weekday] = parts
  
  let desc = ''
  
  if (minute === '*' && hour === '*' && day === '*' && month === '*' && weekday === '*') {
    return '每分钟执行'
  }
  
  if (minute !== '*' && hour === '*' && day === '*' && month === '*' && weekday === '*') {
    return `每小时的第 ${minute} 分钟执行`
  }
  
  if (hour !== '*' && day === '*' && month === '*' && weekday === '*') {
    desc = `每天 ${hour.padStart(2, '0')}:${minute.padStart(2, '0')} 执行`
  } else if (weekday !== '*') {
    const days = weekday.split(',').map(d => weekDays[parseInt(d)] || d).join('、')
    desc = `每周${days} ${hour.padStart(2, '0')}:${minute.padStart(2, '0')} 执行`
  } else if (day !== '*') {
    desc = `每月 ${day} 日 ${hour.padStart(2, '0')}:${minute.padStart(2, '0')} 执行`
  } else {
    desc = `按 cron 表达式执行`
  }
  
  return desc
})

function selectPreset(preset: typeof presets[0]) {
  selectedPreset.value = preset.value
  cronExpression.value = preset.value
  emit('update:modelValue', preset.value)
}

function updateFromCustom() {
  selectedPreset.value = ''
  let expr = ''
  
  switch (customConfig.frequency) {
    case 'minute':
      expr = '* * * * *'
      break
    case 'hour':
      expr = `${customConfig.minute} * * * *`
      break
    case 'day': {
      const [h, m] = customConfig.time.split(':')
      expr = `${parseInt(m)} ${parseInt(h)} * * *`
      break
    }
    case 'week': {
      const [h, m] = customConfig.time.split(':')
      const days = customConfig.weekdays.sort().join(',')
      expr = `${parseInt(m)} ${parseInt(h)} * * ${days}`
      break
    }
    case 'month': {
      const [h, m] = customConfig.time.split(':')
      expr = `${parseInt(m)} ${parseInt(h)} ${customConfig.dayOfMonth} * *`
      break
    }
    case 'interval':
      expr = `@every ${customConfig.intervalValue}${customConfig.intervalUnit}`
      break
  }
  
  cronExpression.value = expr
  emit('update:modelValue', expr)
}

function handleModeChange() {
  // 切换模式时保留表达式
}

function handleAdvancedInput() {
  emit('update:modelValue', cronExpression.value)
}

function copyExpression() {
  navigator.clipboard.writeText(cronExpression.value)
  ElMessage.success('已复制到剪贴板')
}

// 监听外部值变化
watch(() => props.modelValue, (newVal) => {
  if (newVal !== cronExpression.value) {
    cronExpression.value = newVal
    // 检查是否匹配预设
    const matchedPreset = presets.find(p => p.value === newVal)
    if (matchedPreset) {
      selectedPreset.value = matchedPreset.value
    } else {
      selectedPreset.value = ''
    }
  }
})

// 初始化时解析已有值
onMounted(() => {
  if (props.modelValue) {
    cronExpression.value = props.modelValue
    const matchedPreset = presets.find(p => p.value === props.modelValue)
    if (matchedPreset) {
      selectedPreset.value = matchedPreset.value
    }
  }
})
</script>

<style lang="scss" scoped>
.cron-picker {
  .mode-tabs {
    margin-bottom: 16px;
    
    :deep(.el-radio-group) {
      width: 100%;
      display: flex;
      
      .el-radio-button {
        flex: 1;
        
        .el-radio-button__inner {
          width: 100%;
        }
      }
    }
  }
  
  .section-title {
    font-size: 13px;
    color: var(--text-secondary);
    margin-bottom: 12px;
    font-weight: 500;
  }
  
  .simple-mode {
    .preset-section {
      margin-bottom: 20px;
      
      .preset-grid {
        display: grid;
        grid-template-columns: repeat(4, 1fr);
        gap: 8px;
        
        .preset-item {
          display: flex;
          flex-direction: column;
          align-items: center;
          padding: 12px 8px;
          border: 1px solid var(--border-color);
          border-radius: 8px;
          cursor: pointer;
          transition: all 0.2s;
          background: var(--bg-secondary);
          
          &:hover {
            border-color: var(--primary-color);
            background: rgba(var(--primary-color-rgb), 0.05);
          }
          
          &.active {
            border-color: var(--primary-color);
            background: rgba(var(--primary-color-rgb), 0.1);
            
            .el-icon, .preset-label {
              color: var(--primary-color);
            }
          }
          
          .el-icon {
            font-size: 20px;
            color: var(--text-secondary);
            margin-bottom: 6px;
          }
          
          .preset-label {
            font-size: 13px;
            font-weight: 500;
            color: var(--text-primary);
            margin-bottom: 2px;
          }
          
          .preset-desc {
            font-size: 11px;
            color: var(--text-muted);
          }
        }
      }
    }
    
    .custom-section {
      background: var(--bg-secondary);
      border-radius: 8px;
      padding: 16px;
      
      .custom-row {
        display: flex;
        align-items: center;
        margin-bottom: 12px;
        
        &:last-child {
          margin-bottom: 0;
        }
        
        .row-label {
          width: 70px;
          flex-shrink: 0;
          font-size: 13px;
          color: var(--text-secondary);
        }
        
        .el-select, .el-time-picker {
          flex: 1;
        }
        
        .interval-config {
          display: flex;
          align-items: center;
          gap: 8px;
          flex: 1;
          
          span {
            color: var(--text-secondary);
            font-size: 13px;
          }
          
          .el-input-number {
            width: 100px;
          }
          
          .el-select {
            width: 80px;
            flex: none;
          }
        }
        
        :deep(.el-checkbox-group) {
          display: flex;
          flex-wrap: wrap;
          gap: 4px;
          
          .el-checkbox-button {
            .el-checkbox-button__inner {
              padding: 6px 10px;
              border-radius: 4px;
            }
          }
        }
      }
    }
  }
  
  .advanced-mode {
    .cron-input-wrapper {
      margin-bottom: 16px;
      
      :deep(.el-input__wrapper) {
        background: var(--input-bg);
      }
    }
    
    .cron-help {
      background: var(--bg-secondary);
      border-radius: 8px;
      padding: 12px 16px;
      
      .help-title {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: 13px;
        color: var(--text-secondary);
        margin-bottom: 10px;
        
        .el-icon {
          color: var(--info-color);
        }
      }
      
      .help-content {
        .help-row {
          display: flex;
          align-items: center;
          gap: 12px;
          margin-bottom: 6px;
          font-size: 12px;
          
          &:last-child {
            margin-bottom: 0;
          }
          
          code {
            font-family: var(--font-family-mono);
            background: var(--bg-primary);
            padding: 2px 8px;
            border-radius: 4px;
            color: var(--primary-color);
            min-width: 120px;
          }
          
          span {
            color: var(--text-muted);
          }
        }
      }
    }
  }
  
  .result-preview {
    margin-top: 16px;
    padding: 12px 16px;
    background: var(--bg-secondary);
    border-radius: 8px;
    border: 1px solid var(--border-color);
    
    .preview-label {
      font-size: 12px;
      color: var(--text-muted);
      margin-bottom: 6px;
    }
    
    .preview-expression {
      display: flex;
      align-items: center;
      justify-content: space-between;
      
      code {
        font-family: var(--font-family-mono);
        font-size: 16px;
        font-weight: 600;
        color: var(--primary-color);
      }
    }
    
    .preview-description {
      display: flex;
      align-items: center;
      gap: 6px;
      margin-top: 8px;
      font-size: 13px;
      color: var(--text-secondary);
      
      .el-icon {
        color: var(--info-color);
      }
    }
  }
}
</style>
