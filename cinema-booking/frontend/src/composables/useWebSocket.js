import { ref, onUnmounted } from 'vue'

export function useWebSocket(showtimeId) {
  const ws = ref(null)
  const connected = ref(false)
  const handlers = ref([])

  function connect() {
    const token = localStorage.getItem('cinema_token') || ''
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const url = `${protocol}//${host}/ws/showtimes/${showtimeId}?token=${token}`

    ws.value = new WebSocket(url)

    ws.value.onopen = () => { connected.value = true }
    ws.value.onclose = () => {
      connected.value = false
      // Auto-reconnect after 3s
      setTimeout(connect, 3000)
    }
    ws.value.onerror = () => { connected.value = false }
    ws.value.onmessage = (evt) => {
      try {
        const msg = JSON.parse(evt.data)
        handlers.value.forEach(fn => fn(msg))
      } catch {}
    }
  }

  function onMessage(fn) {
    handlers.value.push(fn)
  }

  function disconnect() {
    if (ws.value) {
      ws.value.onclose = null // prevent auto-reconnect
      ws.value.close()
    }
  }

  onUnmounted(disconnect)

  return { connected, connect, onMessage, disconnect }
}
