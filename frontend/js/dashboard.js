document.addEventListener('DOMContentLoaded', async () => {
    try {
        const res = await fetch(`${window.API_URL}/auth/me`, { credentials: 'include' })
        if (!res.ok) throw new Error()
        const data = await res.json()
        document.getElementById('userEmail').textContent = data.email
    } catch {
        window.location.href = './login.html'
        return
    }

    document.getElementById('logoutBtn').addEventListener('click', async () => {
        try {
            await fetch(`${window.API_URL}/auth/logout`, { method: 'POST', credentials: 'include' })
        } catch {}
        window.location.href = './login.html'
    })

    const addBtn = document.getElementById('addBtn')
    const addMsg = document.getElementById('addMessage')

    document.getElementById('addBtn').addEventListener('click', async () => {
        const urlInput = document.getElementById('websiteUrl')
        const website = urlInput.value.trim()
        if (!website) return

        setMsg(addMsg, '')
        setLoading(addBtn, true, 'Adding...')

        try {
            const res = await fetch(`${window.API_URL}/add`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({ website })
            })
            const data = await res.json()

            if (res.ok) {
                urlInput.value = ''
                setMsg(addMsg, data.message || 'Website added.', 'success')
                await loadWebsites()
            } else if (res.status === 401) {
                window.location.href = './login.html'
            } else {
                setMsg(addMsg, data.error || 'Failed to add website', 'error')
            }
        } catch {
            setMsg(addMsg, 'Network error.', 'error')
        } finally {
            setLoading(addBtn, false, 'Add')
        }
    })

    document.getElementById('websiteUrl').addEventListener('keydown', (e) => {
        if (e.key === 'Enter') { e.preventDefault(); addBtn.click() }
    })

    const forceBtn = document.getElementById('forceCheckBtn')
    forceBtn.addEventListener('click', async () => {
        setLoading(forceBtn, true, 'Checking...')
        try {
            const res = await fetch(`${window.API_URL}/force-check`, {
                method: 'POST',
                credentials: 'include'
            })
            const data = await res.json()
            showToast(res.ok ? (data.message || 'Check started') : (data.error || 'Failed'), !res.ok)
            if (res.ok) setTimeout(() => loadWebsites(), 2000)
        } catch {
            showToast('Network error.', true)
        } finally {
            setLoading(forceBtn, false, 'Check all')
        }
    })

    async function loadWebsites(silent = false) {
        const list = document.getElementById('websiteList')
        
        const isDeleting = Array.from(list.querySelectorAll('.btn-danger')).some(btn => btn.disabled)
        if (isDeleting) return

        if (!silent) {
            list.innerHTML = '<li class="state-text">Loading...</li>'
        }

        try {
            const res = await fetch(`${window.API_URL}/websites`, { credentials: 'include' })

            if (res.status === 401) { window.location.href = './login.html'; return }
            if (!res.ok) throw new Error()

            const data = await res.json()
            const sites = data.websites || []

            if (sites.length === 0) {
                list.innerHTML = '<li class="state-text">No websites being tracked yet.</li>'
                return
            }

            const newItems = []
            sites.forEach(({ url, status }) => {
                const statusClass = status === 'UP' ? 'up' : status === 'DOWN' ? 'down' : 'pending'

                const li = document.createElement('li')
                li.className = 'site-item'

                const dot = document.createElement('span')
                dot.className = `status-dot ${statusClass}`

                const urlSpan = document.createElement('span')
                urlSpan.className = 'site-url'
                urlSpan.textContent = url
                urlSpan.title = url

                const badge = document.createElement('span')
                badge.className = `site-badge ${statusClass}`
                badge.textContent = status

                const left = document.createElement('div')
                left.className = 'site-left'
                left.append(dot, urlSpan, badge)

                const delBtn = document.createElement('button')
                delBtn.className = 'btn btn-danger'
                delBtn.textContent = 'Remove'
                delBtn.addEventListener('click', () => deleteWebsite(url, li, delBtn))

                li.append(left, delBtn)
                newItems.push(li)
            })

            list.replaceChildren(...newItems)
        } catch {
            if (!silent) {
                list.innerHTML = '<li class="state-text">Failed to load websites.</li>'
            }
        }
    }

    async function deleteWebsite(url, li, btn) {
        if (!window.confirm(`Stop tracking ${url}?`)) return
        setLoading(btn, true, 'Removing...')
        try {
            const res = await fetch(`${window.API_URL}/deleteweb?website=${encodeURIComponent(url)}`, {
                method: 'DELETE',
                credentials: 'include'
            })
            if (res.ok) {
                li.remove()
                const list = document.getElementById('websiteList')
                if (!list.children.length) {
                    list.innerHTML = '<li class="state-text">No websites being tracked yet.</li>'
                }
                showToast('Website removed.')
            } else {
                const data = await res.json()
                showToast(data.error || 'Failed to remove', true)
                setLoading(btn, false, 'Remove')
            }
        } catch {
            showToast('Network error.', true)
            setLoading(btn, false, 'Remove')
        }
    }

    loadWebsites()

    setInterval(() => loadWebsites(true), 5000)
})

function setMsg(el, text, type = '') {
    el.textContent = text
    el.className = 'msg' + (type ? ' ' + type : '')
}

function setLoading(btn, loading, label) {
    btn.disabled = loading
    btn.textContent = label
}

function showToast(message, isError = false) {
    const toast = document.getElementById('toast')
    toast.textContent = message
    toast.className = 'show' + (isError ? ' toast-error' : '')
    clearTimeout(toast._timer)
    toast._timer = setTimeout(() => { toast.className = '' }, 3000)
}
