document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('forgotForm')
    const msg = document.getElementById('authMessage')
    const btn = document.getElementById('submitBtn')

    form.addEventListener('submit', async (e) => {
        e.preventDefault()
        setMsg(msg, '')
        setLoading(btn, true, 'Sending...')

        const email = document.getElementById('email').value.trim()

        try {
            const res = await fetch(`${window.API_URL}/auth/forgot-password`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ email })
            })
            const data = await res.json()

            if (res.ok) {
                setMsg(msg, data.message || 'If this email exists, a reset link has been sent.', 'success')
                form.reset()
            } else {
                setMsg(msg, data.error || 'Something went wrong', 'error')
            }
        } catch {
            setMsg(msg, 'Network error. Is the server running?', 'error')
        } finally {
            setLoading(btn, false, 'Send reset link')
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
