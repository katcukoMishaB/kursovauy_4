// Модуль авторизации и управления пользователями

function checkAuth() {
    const API_BASE = window.AppConfig.API_BASE;
    const token = localStorage.getItem('token');
    if (token) {
        fetch(`${API_BASE}/users/profile`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        })
        .then(res => {
            if (res.ok) {
                return res.json();
            }
            throw new Error('Not authenticated');
        })
        .then(user => {
            window.AppConfig.setCurrentUser(user);
            window.AppConfig.setCurrentRole(localStorage.getItem('role'));
            window.Utils.updateUI();
            window.Utils.showPage('projects');
        })
        .catch(() => {
            localStorage.removeItem('token');
            localStorage.removeItem('role');
            window.Utils.showPage('home');
        });
    } else {
        window.Utils.showPage('home');
    }
}

function login(event) {
    event.preventDefault();
    const API_BASE = window.AppConfig.API_BASE;
    const email = document.getElementById('login-email').value;
    const password = document.getElementById('login-password').value;

    fetch(`${API_BASE}/users/login`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ email, password })
    })
    .then(res => {
        return res.json().then(data => {
            if (!res.ok) {
                throw new Error(data.error || 'Ошибка входа');
            }
            return data;
        });
    })
    .then(data => {
        if (data.token) {
            localStorage.setItem('token', data.token);
            localStorage.setItem('role', data.role);
            window.AppConfig.setCurrentUser(data.user);
            window.AppConfig.setCurrentRole(data.role);
            window.Utils.updateUI();
            window.Utils.showPage('projects');
        } else {
            window.Utils.showMessage('Ошибка входа', 'error');
        }
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка входа: ' + err.message, 'error');
    });
}

function register(event) {
    event.preventDefault();
    const API_BASE = window.AppConfig.API_BASE;
    const firstName = document.getElementById('reg-first-name').value.trim();
    const lastName = document.getElementById('reg-last-name').value.trim();
    const email = document.getElementById('reg-email').value.trim();
    const phone = document.getElementById('reg-phone').value.trim();
    const password = document.getElementById('reg-password').value;
    
    if (!firstName || !lastName || !email || !phone || !password) {
        window.Utils.showMessage('Все поля обязательны для заполнения', 'error');
        return;
    }
    
    const data = {
        first_name: firstName,
        last_name: lastName,
        email: email,
        phone: phone,
        password: password
    };

    fetch(`${API_BASE}/users/register`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    })
    .then(res => {
        return res.json().then(data => {
            if (!res.ok) {
                throw new Error(data.error || 'Ошибка регистрации');
            }
            return data;
        });
    })
    .then(data => {
        if (data.id) {
            window.Utils.showMessage('Регистрация успешна', 'success');
            setTimeout(() => window.Utils.showPage('home'), 2000);
        } else {
            window.Utils.showMessage('Ошибка регистрации', 'error');
        }
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка регистрации: ' + err.message, 'error');
    });
}

function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('role');
    window.AppConfig.setCurrentUser(null);
    window.AppConfig.setCurrentRole(null);
    window.Utils.updateUI();
    window.Utils.showPage('home');
}

function loadProfile() {
    const API_BASE = window.AppConfig.API_BASE;
    const currentRole = window.AppConfig.getCurrentRole();
    fetch(`${API_BASE}/users/profile`, {
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
    })
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка загрузки профиля');
            });
        }
        return res.json();
    })
    .then(user => {
        const content = document.getElementById('profile-content');
        const avatarText = document.getElementById('profile-avatar-text');
        const profileName = document.getElementById('profile-name');
        const profileRole = document.getElementById('profile-role');
        const isOrganizer = currentRole === 'organizer' || currentRole === 'admin';
        
        if (avatarText) {
            avatarText.textContent = (user.first_name || 'У')[0].toUpperCase();
        }
        if (profileName) {
            profileName.textContent = `${window.Utils.escapeHtml(user.first_name)} ${window.Utils.escapeHtml(user.last_name)}`;
        }
        if (profileRole) {
            const roleNames = {
                'admin': 'Администратор',
                'organizer': 'Организатор',
                'participant': 'Участник'
            };
            profileRole.textContent = roleNames[currentRole] || currentRole;
        }
        
        content.innerHTML = `
            <div id="profile-display">
                <div class="profile-field">
                    <label>Имя</label>
                    <div class="profile-field-value">${window.Utils.escapeHtml(user.first_name)}</div>
                </div>
                <div class="profile-field">
                    <label>Фамилия</label>
                    <div class="profile-field-value">${window.Utils.escapeHtml(user.last_name)}</div>
                </div>
                <div class="profile-field">
                    <label>Email</label>
                    <div class="profile-field-value">${window.Utils.escapeHtml(user.email)}</div>
                </div>
                <div class="profile-field">
                    <label>Телефон</label>
                    <div class="profile-field-value">${window.Utils.escapeHtml(user.phone || 'Не указан')}</div>
                </div>
                <div class="profile-field">
                    <label>Роль</label>
                    <div class="profile-field-value">
                        <span class="status ${currentRole}">${currentRole}</span>
                    </div>
                </div>
            </div>
            <div id="profile-edit" style="display:none;">
                <form onsubmit="window.saveProfile(event)">
                    <div class="profile-field">
                        <label>Имя</label>
                        <input type="text" id="edit-first-name" placeholder="Имя" value="${window.Utils.escapeHtml(user.first_name)}" required>
                    </div>
                    <div class="profile-field">
                        <label>Фамилия</label>
                        <input type="text" id="edit-last-name" placeholder="Фамилия" value="${window.Utils.escapeHtml(user.last_name)}" required>
                    </div>
                    <div class="profile-field">
                        <label>Телефон</label>
                        <input type="tel" id="edit-phone" placeholder="Телефон" value="${window.Utils.escapeHtml(user.phone || '')}" required>
                    </div>
                    <div style="display: flex; gap: var(--spacing); margin-top: var(--spacing-lg);">
                        <button type="submit">Сохранить</button>
                        <button type="button" onclick="window.cancelEditProfile()" class="secondary">Отмена</button>
                    </div>
                </form>
            </div>
        `;
        
        // Показываем заявку на роль организатора только для участников (не организаторов и не админов)
        if (currentRole === 'participant') {
            const section = document.getElementById('organizer-request-section');
            if (section) section.style.display = 'block';
        } else {
            const section = document.getElementById('organizer-request-section');
            if (section) section.style.display = 'none';
        }
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка загрузки профиля: ' + err.message, 'error');
    });
}

// Функции для редактирования профиля
window.editProfile = function() {
    document.getElementById('profile-display').style.display = 'none';
    document.getElementById('profile-edit').style.display = 'block';
};

window.cancelEditProfile = function() {
    document.getElementById('profile-display').style.display = 'block';
    document.getElementById('profile-edit').style.display = 'none';
    loadProfile();
};

window.saveProfile = function(event) {
    event.preventDefault();
    const API_BASE = window.AppConfig.API_BASE;
    const firstName = document.getElementById('edit-first-name').value;
    const lastName = document.getElementById('edit-last-name').value;
    const phone = document.getElementById('edit-phone').value;
    
    if (!firstName || !lastName || !phone) {
        window.Utils.showMessage('Все поля обязательны для заполнения', 'error');
        return;
    }
    
    fetch(`${API_BASE}/users/profile`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({
            first_name: firstName,
            last_name: lastName,
            phone: phone
        })
    })
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка обновления профиля');
            });
        }
        return res.json();
    })
    .then(() => {
        window.Utils.showMessage('Профиль обновлен', 'success');
        loadProfile();
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка обновления профиля: ' + err.message, 'error');
    });
};

function submitOrganizerRequest(event) {
    event.preventDefault();
    const API_BASE = window.AppConfig.API_BASE;
    const experience = document.getElementById('organizer-experience').value;

    if (!experience || experience.trim() === '') {
        window.Utils.showMessage('Пожалуйста, заполните описание опыта', 'error');
        return;
    }

    fetch(`${API_BASE}/users/organizer-request`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({ experience_description: experience })
    })
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка подачи заявки');
            });
        }
        return res.json();
    })
    .then(data => {
        if (data.id || data.message) {
            window.Utils.showMessage('Заявка подана успешно', 'success');
            document.getElementById('organizer-request-section').style.display = 'none';
            document.getElementById('organizer-experience').value = '';
        }
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка подачи заявки: ' + err.message, 'error');
    });
}

// Экспорт функций
window.Auth = {
    checkAuth,
    login,
    register,
    logout,
    loadProfile,
    submitOrganizerRequest
};

