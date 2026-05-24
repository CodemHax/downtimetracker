document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('resetForm')
    const msg = document.getElementById('authMessage')
    const btn = document.getElementById('submitBtn')
    const emailInput = document.getElementById('email')
    const tokenInput = document.getElementById('token')

    const urlParams = new URLSearchParams(window.location.search)
    const email = urlParams.get('email')
    const token = urlParams.get('token')

    if (!email || !token) {
        setMsg(msg, 'Invalid or missing reset link. Please request a new one.', 'error')
        form.style.display = 'none'
        return
    }

    emailInput.value = email
    tokenInput.value = token

    form.addEventListener('submit', async (e) => {
        e.preventDefault()
        setMsg(msg, '')

        const password = document.getElementById('password').value
        const confirmPassword = document.getElementById('confirmPassword').value

        if (password.length < 6) {
            setMsg(msg, 'Password must be at least 6 characters.', 'error')
            return
        }

        if (password !== confirmPassword) {
            setMsg(msg, 'Passwords do not match.', 'error')
            return
        }

        setLoading(btn, true, 'Resetting...')

        try {
            const res = await fetch(`${window.API_URL}/auth/reset-password`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    email: emailInput.value,
                    token: tokenInput.value,
                    password
                })
            })
            const data = await res.json()

            if (res.ok) {
                setMsg(msg, data.message || 'Password reset. Redirecting to login...', 'success')
                form.reset()
                setTimeout(() => { window.location.href = './login.html' }, 2500)
            } else {
                setMsg(msg, data.error || 'Reset failed. The link may have expired.', 'error')
            }
        } catch {
            setMsg(msg, 'Network error. Is the server running?', 'error')
        } finally {
            setLoading(btn, false, 'Reset password')
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
