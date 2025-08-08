<template>
  <div ref="chartContainer" class="w-full h-64"></div>
</template>

<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'

interface ChartData {
  date: string
  successful: number
  failed: number
}

interface Props {
  data: ChartData[]
}

const props = defineProps<Props>()
const chartContainer = ref<HTMLElement>()

let chart: any = null

const renderChart = async () => {
  if (!chartContainer.value || !process.client) return

  // Dynamically import ApexCharts on client side only
  const { default: ApexCharts } = await import('apexcharts')

  const options = {
    chart: {
      type: 'bar',
      height: 256,
      toolbar: {
        show: false
      },
      background: 'transparent'
    },
    series: [
      {
        name: 'Successful',
        data: props.data.map(d => d.successful)
      },
      {
        name: 'Failed',
        data: props.data.map(d => d.failed)
      }
    ],
    xaxis: {
      categories: props.data.map(d => {
        const date = new Date(d.date)
        return date.toLocaleDateString('en', { month: 'short', day: 'numeric' })
      }),
      labels: {
        style: {
          colors: '#9CA3AF'
        }
      }
    },
    yaxis: {
      labels: {
        style: {
          colors: '#9CA3AF'
        }
      }
    },
    colors: ['#10B981', '#EF4444'],
    plotOptions: {
      bar: {
        borderRadius: 4,
        columnWidth: '60%'
      }
    },
    dataLabels: {
      enabled: false
    },
    legend: {
      position: 'top',
      horizontalAlign: 'right',
      labels: {
        colors: '#9CA3AF'
      }
    },
    grid: {
      borderColor: '#374151',
      strokeDashArray: 4
    },
    theme: {
      mode: 'dark'
    }
  }

  if (chart) {
    chart.updateOptions(options)
  } else {
    chart = new ApexCharts(chartContainer.value, options)
    await chart.render()
  }
}

onMounted(() => {
  renderChart()
})

watch(() => props.data, () => {
  renderChart()
}, { deep: true })

onUnmounted(() => {
  if (chart) {
    chart.destroy()
  }
})
</script>