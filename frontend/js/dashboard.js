

function getHeaders() {
    return {
        'Content-Type': 'application/json'
    }
}

document.addEventListener('DOMContentLoaded', async () => {

    try {
        const res = await fetch(`${window.API_URL}/auth/me`, {
            credentials: 'include'
        })
        if (!res.ok) throw new Error()
        const data = await res.json()
        document.getElementById('userEmail').innerText = data.email
    } catch (e) {
        window.location.href = './login.html'
        return
    }

    document.getElementById('logoutBtn').addEventListener('click', async () => {
        try {
            await fetch(`${window.API_URL}/auth/logout`, { method: 'POST', credentials: 'include' })
        } catch(e) {}
        localStorage.removeItem('email')
        window.location.href = './login.html'
    })

    const form = document.getElementById('addWebForm')
    const addMessage = document.getElementById('addMessage')
    const forceCheckBtn = document.getElementById('forceCheckBtn')

    form.addEventListener('submit', async (e) => {
        e.preventDefault()
        addMessage.innerText = 'Adding...'
        addMessage.className = 'message'
        const website = document.getElementById('websiteUrl').value

        try {
            const res = await fetch(`${window.API_URL}/add`, {
                method: 'POST',
                headers: getHeaders(),
                credentials: 'include',
                body: JSON.stringify({ website })
            })
            const data = await res.json()
            if (res.ok) {
                addMessage.innerText = data.message
                addMessage.classList.add('success')
                document.getElementById('websiteUrl').value = ''
                loadWebsites()
            } else {
                throw new Error(data.error || 'Failed to add website')
            }
        } catch (error) {
            if (error.message.includes('401') || error.message.includes('login')) window.location.href = './login.html'
            addMessage.innerText = error.message
            addMessage.classList.add('error')
        }
    })

    forceCheckBtn.addEventListener('click', async () => {
        try {
            const res = await fetch(`${window.API_URL}/force-check`, {
                method: 'POST',
                headers: getHeaders(),
                credentials: 'include'
            })
            const data = await res.json()
            if (res.ok) {
                alert(data.message)
            } else {
                alert(data.error)
            }
        } catch (error) {
            alert('Check failed: ' + error.message)
        }
    })

    window.deleteWebsite = async (website) => {
        if (!confirm(`Are you sure you want to stop tracking ${website}?`)) return
        try {
            const res = await fetch(`${window.API_URL}/deleteweb?website=${encodeURIComponent(website)}`, {
                method: 'DELETE',
                headers: getHeaders(),
                credentials: 'include'
            })
            if (res.ok) {
                loadWebsites()
            } else {
                const data = await res.json()
                alert(data.error || 'Failed to delete')
            }
        } catch(error) {
            alert('Error deleting: ' + error.message)
        }
    }

    async function loadWebsites() {
        const list = document.getElementById('websiteList')
        list.innerHTML = '<p style="color:var(--text-muted)">Loading tracking list...</p>'
        try {
            const res = await fetch(`${window.API_URL}/websites`, {
                headers: getHeaders(),
                credentials: 'include'
            })
            if (res.ok) {
                const data = await res.json()
                const sites = data.websites || []
                list.innerHTML = sites.length ? '' : '<p style="color:var(--text-muted)">No websites being tracked.</p>'
                
                sites.forEach(siteObj => {
                    const site = siteObj.url;
                    const status = siteObj.status;
                    let statusColor = 'var(--text-muted)';
                    if (status === 'UP') statusColor = 'var(--success-color)';
                    if (status === 'DOWN') statusColor = 'var(--danger-color)';
                    
                    const li = document.createElement('li')
                    li.className = 'website-item'
                    li.innerHTML = `
                        <div class="site-info">
                            <span class="site-url">${site}</span>
                            <span class="site-status" style="color: ${statusColor};">${status}</span>
                        </div>
                        <button onclick="deleteWebsite('${site}')" class="btn danger">Remove</button>
                    `
                    list.appendChild(li)
                })
            } else if (res.status === 401) {
                window.location.href = './login.html'
            }
        } catch (err) {
            list.innerHTML = '<p class="error">Failed to load websites</p>'
        }
    }

    loadWebsites()
})
