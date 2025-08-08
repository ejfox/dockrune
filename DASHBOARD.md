# ⚡ dockrune Dashboard - LIGHTNING FAST!

A minimal but delightful admin dashboard for dockrune deployment management.

## 🚀 Quick Start

```bash
# Install and run - WHOOSH! 
python3 run.py

# Or manually:
pip install -r requirements.txt
python3 main.py
```

Then visit: **http://localhost:8000/admin**

## ✨ Features

- **Real-time status updates** via WebSocket - *ZOOM!*
- **Smooth anime.js animations** - *SWOOSH!*  
- **Live app monitoring** with status indicators that GLOW!
- **One-click deploy/stop** with visual feedback
- **Smooth-scrolling log viewer** - buttery smooth!
- **Responsive design** - works on all devices
- **"Ugly but honest"** aesthetic with delightful interactions

## 🎯 Status Indicators

- 🟢 **Live** - App running smoothly
- 🟡 **Deploying** - Deployment in progress (pulsing!)
- 🔴 **Failing** - Something went wrong
- ⚪ **Stopped** - App is stopped

## 🎮 Interactions

- **Deploy Button** - Triggers redeployment with smooth animation
- **Stop Button** - Stops the app with visual feedback  
- **Logs Button** - Opens smooth modal with scrollable logs
- **App Cards** - Hover for smooth elevation effects
- **Status Badges** - Pulse when deploying for that *WHOOSH* factor!

## 🛠 API Endpoints

- `GET /api/apps` - List all apps
- `GET /api/apps/{name}` - Get specific app status
- `GET /api/apps/{name}/logs` - Get app logs
- `POST /api/apps/{name}/deploy` - Redeploy app
- `POST /api/apps/{name}/stop` - Stop app
- `WS /ws` - Real-time status updates

## 🎨 Tech Stack

- **Backend**: FastAPI + WebSockets - LIGHTNING FAST! ⚡
- **Frontend**: Vue 3 Composition API - REACTIVE POWER!
- **Animations**: anime.js - SMOOTH AS SILK!
- **Styling**: Custom CSS with CSS Grid - RESPONSIVE MAGIC!

## 🔥 Performance Features

- Staggered animations on load - *ZOOM ZOOM*
- Optimized WebSocket updates
- Efficient DOM updates with Vue reactivity
- Smooth 60fps animations with anime.js
- Responsive grid that adapts like LIGHTNING!

---

*Built by SPEED DEMON ⚡ - "Making deployment dashboards fast and delightful!"*