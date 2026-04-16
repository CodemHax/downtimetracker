

document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('registerForm')
    const authMessage = document.getElementById('authMessage')

    form.addEventListener('submit', async (e) => {
        e.preventDefault()
        authMessage.innerText = 'Registring...'
        authMessage.className = 'message'
        
        const email = document.getElementById('email').value
        const password = document.getElementById('password').value

        try {
            const res = await fetch(`${API_URL}/auth/register`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                credentials: 'include',
                body: JSON.stringify({ email, password })
            })
            const data = await res.json()
            
            if (res.ok) {
                authMessage.innerText = data.message
                authMessage.classList.add('success')
                form.reset()
            } else {
                throw new Error(data.error || 'Registration failed')
            }
        } catch (error) {
            authMessage.innerText = error.message
            authMessage.classList.add('error')
        }
    })
})
