document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('loginForm')
    const msg = document.getElementById('authMessage')
    const btn = document.getElementById('submitBtn')

    form.addEventListener('submit', async (e) => {
        e.preventDefault()
        setMsg(msg, '')
        setLoading(btn, true, 'Signing in...')

        const email = document.getElementById('email').value.trim()
        const password = document.getElementById('password').value

        try {
            const res = await fetch(`${window.API_URL}/auth/login`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({ email, password })
            })
            const data = await res.json()

            if (res.ok) {
                window.location.href = './index.html'
            } else {
                setMsg(msg, data.error || 'Login failed', 'error')
            }
        } catch {
            setMsg(msg, 'Network error. Is the server running?', 'error')
        } finally {
            setLoading(btn, false, 'Sign in')
        }
    })
})

function setMsg(el, text, type = '') {
    el.textContent = text
    el.className = 'msg' + (type ? ' ' + type : '')
}

function setLoading(btn, loading, label) {
    btn.disabled = loading
    btn.textContent = label
}
