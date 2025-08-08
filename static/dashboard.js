// ⚡ SPEED DEMON'S Vue 3 Dashboard - PREPARE FOR LIGHTNING! 
const { createApp, ref, reactive, onMounted, onUnmounted, computed, nextTick } = Vue;

// This is where the magic happens! ⚡
const Dashboard = {
  setup() {
    // Reactive state - BLAZING FAST! 🏃‍♂️
    const apps = ref([]);
    const selectedApp = ref(null);
    const isConnected = ref(false);
    const lastUpdate = ref(new Date());
    const loading = ref(true);
    
    // WebSocket connection - REAL-TIME POWER! ⚡
    let websocket = null;
    
    // Status color mapping - these will GLOW! ✨
    const statusColors = {
      live: '#10b981',      // Green - smooth and healthy
      deploying: '#f59e0b', // Amber - pulsing with energy
      failing: '#ef4444',   // Red - attention-grabbing
      stopped: '#6b7280'    // Gray - calm and collected
    };
    
    // Connect to WebSocket - WHOOSH! 🌪️
    const connectWebSocket = () => {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
      const wsUrl = `${protocol}//${window.location.host}/ws`;
      
      websocket = new WebSocket(wsUrl);
      
      websocket.onopen = () => {
        isConnected.value = true;
        console.log('⚡ WebSocket connected - LIGHTNING SPEED!');
        
        // Animate connection indicator - *SWOOSH*
        anime({
          targets: '.connection-indicator',
          scale: [1, 1.2, 1],
          duration: 600,
          easing: 'easeOutElastic(1, .8)'
        });
      };
      
      websocket.onmessage = (event) => {
        const message = JSON.parse(event.data);
        
        if (message.type === 'init') {
          apps.value = message.data;
          loading.value = false;
          
          // Stagger animation for initial load - *ZOOM ZOOM*
          nextTick(() => {
            anime({
              targets: '.app-card',
              translateY: [50, 0],
              opacity: [0, 1],
              delay: anime.stagger(100),
              duration: 800,
              easing: 'easeOutCubic'
            });
          });
          
        } else if (message.type === 'status_update') {
          updateAppStatus(message.app_name, message.data);
        }
        
        lastUpdate.value = new Date();
      };
      
      websocket.onclose = () => {
        isConnected.value = false;
        console.log('WebSocket disconnected - attempting reconnect...');
        setTimeout(connectWebSocket, 3000);
      };
    };
    
    // Update app status with smooth animations - *WHOOSH*
    const updateAppStatus = (appName, newData) => {
      const appIndex = apps.value.findIndex(app => app.name === appName);
      if (appIndex !== -1) {
        const oldStatus = apps.value[appIndex].status;
        apps.value[appIndex] = newData;
        
        // If status changed, animate the card! - *SWOOSH*
        if (oldStatus !== newData.status) {
          nextTick(() => {
            const cardElement = document.querySelector(`[data-app="${appName}"]`);
            if (cardElement) {
              anime({
                targets: cardElement,
                scale: [1, 1.05, 1],
                duration: 400,
                easing: 'easeOutElastic(1, .8)'
              });
              
              // Pulse the status indicator - *GLOW*
              const statusEl = cardElement.querySelector('.status-indicator');
              if (statusEl) {
                anime({
                  targets: statusEl,
                  boxShadow: [
                    `0 0 0px ${statusColors[newData.status]}`,
                    `0 0 20px ${statusColors[newData.status]}`,
                    `0 0 0px ${statusColors[newData.status]}`
                  ],
                  duration: 1000,
                  easing: 'easeInOutQuad'
                });
              }
            }
          });
        }
      }
    };
    
    // Deploy action with visual feedback - LIGHTNING FAST! ⚡
    const deployApp = async (appName) => {
      const button = document.querySelector(`[data-deploy="${appName}"]`);
      
      // Button animation - *WHOOSH*
      anime({
        targets: button,
        scale: [1, 0.95, 1],
        duration: 200,
        easing: 'easeOutQuad'
      });
      
      try {
        const response = await fetch(`/api/apps/${appName}/deploy`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ app_name: appName })
        });
        
        if (response.ok) {
          // Success animation - *SPARKLE*
          anime({
            targets: button,
            backgroundColor: ['#10b981', '#059669', '#10b981'],
            duration: 1000,
            easing: 'easeInOutQuad'
          });
        }
      } catch (error) {
        console.error('Deploy failed:', error);
        // Error animation - *SHAKE*
        anime({
          targets: button,
          translateX: [-10, 10, -5, 5, 0],
          duration: 400,
          easing: 'easeInOutQuad'
        });
      }
    };
    
    // Stop action - clean and swift! 🛑
    const stopApp = async (appName) => {
      const button = document.querySelector(`[data-stop="${appName}"]`);
      
      anime({
        targets: button,
        scale: [1, 0.95, 1],
        duration: 200,
        easing: 'easeOutQuad'
      });
      
      try {
        const response = await fetch(`/api/apps/${appName}/stop`, {
          method: 'POST'
        });
        
        if (response.ok) {
          anime({
            targets: button,
            backgroundColor: ['#ef4444', '#dc2626', '#ef4444'],
            duration: 1000,
            easing: 'easeInOutQuad'
          });
        }
      } catch (error) {
        console.error('Stop failed:', error);
      }
    };
    
    // Show app logs - smooth as silk! 🎭
    const showLogs = (app) => {
      selectedApp.value = app;
      
      nextTick(() => {
        const modal = document.querySelector('.log-modal');
        const backdrop = document.querySelector('.modal-backdrop');
        
        // Modal entrance animation - *SWOOSH*
        anime.timeline()
          .add({
            targets: backdrop,
            opacity: [0, 1],
            duration: 200,
            easing: 'easeOutQuad'
          })
          .add({
            targets: modal,
            scale: [0.8, 1],
            opacity: [0, 1],
            duration: 300,
            easing: 'easeOutCubic'
          }, '-=100');
        
        // Auto-scroll to bottom with smooth animation
        const logContainer = document.querySelector('.log-content');
        if (logContainer) {
          anime({
            targets: logContainer,
            scrollTop: logContainer.scrollHeight,
            duration: 800,
            easing: 'easeOutCubic'
          });
        }
      });
    };
    
    // Close modal - *ZOOM OUT*
    const closeLogs = () => {
      const modal = document.querySelector('.log-modal');
      const backdrop = document.querySelector('.modal-backdrop');
      
      anime.timeline()
        .add({
          targets: modal,
          scale: [1, 0.8],
          opacity: [1, 0],
          duration: 200,
          easing: 'easeInQuad'
        })
        .add({
          targets: backdrop,
          opacity: [1, 0],
          duration: 200,
          easing: 'easeInQuad',
          complete: () => {
            selectedApp.value = null;
          }
        }, '-=100');
    };
    
    // Format timestamp - clean and readable
    const formatTime = (timestamp) => {
      return new Date(timestamp).toLocaleString();
    };
    
    // Get status display text
    const getStatusText = (status) => {
      const statusMap = {
        live: '🟢 Live',
        deploying: '🟡 Deploying',
        failing: '🔴 Failing', 
        stopped: '⚪ Stopped'
      };
      return statusMap[status] || status;
    };
    
    // Computed properties for filtering and sorting
    const liveApps = computed(() => apps.value.filter(app => app.status === 'live'));
    const deployingApps = computed(() => apps.value.filter(app => app.status === 'deploying'));
    const failingApps = computed(() => apps.value.filter(app => app.status === 'failing'));
    
    // Lifecycle hooks
    onMounted(() => {
      connectWebSocket();
      
      // Continuous pulse animation for deploying apps - *PULSE*
      setInterval(() => {
        const deployingElements = document.querySelectorAll('.status-deploying');
        if (deployingElements.length > 0) {
          anime({
            targets: deployingElements,
            opacity: [1, 0.6, 1],
            duration: 2000,
            easing: 'easeInOutSine'
          });
        }
      }, 2000);
    });
    
    onUnmounted(() => {
      if (websocket) {
        websocket.close();
      }
    });
    
    return {
      apps,
      selectedApp,
      isConnected,
      lastUpdate,
      loading,
      liveApps,
      deployingApps,
      failingApps,
      deployApp,
      stopApp,
      showLogs,
      closeLogs,
      formatTime,
      getStatusText,
      statusColors
    };
  },
  
  template: `
    <div class="dashboard">
      <!-- Header with connection status - SLEEK! -->
      <header class="dashboard-header">
        <div class="header-content">
          <h1 class="dashboard-title">
            ⚡ dockrune admin
            <span class="subtitle">ugly but honest</span>
          </h1>
          <div class="connection-status">
            <div class="connection-indicator" :class="{ connected: isConnected }"></div>
            <span>{{ isConnected ? 'Connected' : 'Disconnected' }}</span>
            <div class="last-update">
              Last update: {{ formatTime(lastUpdate) }}
            </div>
          </div>
        </div>
      </header>
      
      <!-- Loading state - SMOOTH! -->
      <div v-if="loading" class="loading">
        <div class="loading-spinner"></div>
        <p>Loading apps... ⚡</p>
      </div>
      
      <!-- Main dashboard content -->
      <main v-else class="dashboard-content">
        <!-- Quick stats -->
        <div class="stats-row">
          <div class="stat-card live">
            <div class="stat-number">{{ liveApps.length }}</div>
            <div class="stat-label">Live Apps</div>
          </div>
          <div class="stat-card deploying">
            <div class="stat-number">{{ deployingApps.length }}</div>
            <div class="stat-label">Deploying</div>
          </div>
          <div class="stat-card failing">
            <div class="stat-number">{{ failingApps.length }}</div>
            <div class="stat-label">Failing</div>
          </div>
        </div>
        
        <!-- Apps grid - this is where the magic happens! ⚡ -->
        <div class="apps-grid">
          <div 
            v-for="app in apps" 
            :key="app.name"
            class="app-card"
            :class="[\`status-\${app.status}\`]"
            :data-app="app.name"
          >
            <div class="app-header">
              <h3 class="app-name">{{ app.name }}</h3>
              <div class="status-badge" :class="app.status">
                <div class="status-indicator" :style="{ backgroundColor: statusColors[app.status] }"></div>
                {{ getStatusText(app.status) }}
              </div>
            </div>
            
            <div class="app-details">
              <div class="detail-row">
                <span class="label">Domain:</span>
                <span class="value">{{ app.domain }}</span>
              </div>
              <div class="detail-row">
                <span class="label">Port:</span>
                <span class="value">{{ app.port }}</span>
              </div>
              <div class="detail-row">
                <span class="label">SHA:</span>
                <span class="value mono">{{ app.last_deploy_sha }}</span>
              </div>
              <div class="detail-row">
                <span class="label">Last Deploy:</span>
                <span class="value">{{ formatTime(app.last_deploy_timestamp) }}</span>
              </div>
            </div>
            
            <div class="app-actions">
              <button 
                @click="deployApp(app.name)"
                :data-deploy="app.name"
                class="btn btn-deploy"
                :disabled="app.status === 'deploying'"
              >
                {{ app.status === 'deploying' ? 'Deploying...' : '🚀 Deploy' }}
              </button>
              
              <button 
                @click="stopApp(app.name)"
                :data-stop="app.name"
                class="btn btn-stop"
                :disabled="app.status === 'stopped'"
              >
                🛑 Stop
              </button>
              
              <button 
                @click="showLogs(app)"
                class="btn btn-logs"
              >
                📋 Logs
              </button>
            </div>
          </div>
        </div>
      </main>
      
      <!-- Log modal - SMOOTH AS BUTTER! -->
      <div v-if="selectedApp" class="modal-backdrop" @click="closeLogs">
        <div class="log-modal" @click.stop>
          <div class="modal-header">
            <h3>{{ selectedApp.name }} Logs</h3>
            <button @click="closeLogs" class="close-btn">✕</button>
          </div>
          <div class="log-content">
            <div 
              v-for="(line, index) in selectedApp.log_lines" 
              :key="index"
              class="log-line"
            >
              {{ line }}
            </div>
          </div>
        </div>
      </div>
    </div>
  `
};

// Mount the app - LAUNCH! 🚀
createApp(Dashboard).mount('#app');