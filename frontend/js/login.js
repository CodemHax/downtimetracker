

document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('loginForm')
    const authMessage = document.getElementById('authMessage')

    form.addEventListener('submit', async (e) => {
        e.preventDefault()
        authMessage.innerText = 'Authenticating...'
        authMessage.className = 'message'
        
        const email = document.getElementById('email').value
        const password = document.getElementById('password').value

        try {
            const res = await fetch(`${API_URL}/auth/login`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({ email, password })
            })
            const data = await res.json()
            
            if (res.ok) {
                window.location.href = './index.html'
            } else {
                throw new Error(data.error || 'Failed to login')
            }
        } catch (error) {
            authMessage.innerText = error.message
            authMessage.classList.add('error')
        }
    })
})
