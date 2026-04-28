const API_BASE = '/api';

let currentUser = null;
let currentRole = null;

const accessibilitySettings = {
    fontSize: 16,
    highContrast: false,
    simpleNav: false
};

function init() {
    loadAccessibilitySettings();
    applyAccessibilitySettings();
    checkAuth();
    loadCategories();
}

function checkAuth() {
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
            currentUser = user;
            currentRole = localStorage.getItem('role');
            updateUI();
        })
        .catch(() => {
            localStorage.removeItem('token');
            localStorage.removeItem('role');
            showPage('home');
        });
    }
}

function updateUI() {
    if (currentUser) {
        document.getElementById('logout-btn').style.display = 'block';
        document.getElementById('profile-link').style.display = 'block';
        if (currentRole === 'organizer' || currentRole === 'admin') {
            document.getElementById('create-project-btn').style.display = 'block';
        }
        if (currentRole === 'admin') {
            document.getElementById('admin-link').style.display = 'block';
        }
    }
}

function showPage(pageId) {
    document.querySelectorAll('.page').forEach(page => {
        page.style.display = 'none';
    });
    document.getElementById(`${pageId}-page`).style.display = 'block';
    
    if (pageId === 'projects') {
        loadProjects();
    } else if (pageId === 'profile') {
        loadProfile();
    } else if (pageId === 'admin') {
        showAdminTab('users');
    }
}

function login(event) {
    event.preventDefault();
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
            currentUser = data.user;
            currentRole = data.role;
            updateUI();
            showPage('projects');
            showMessage('Вход выполнен успешно', 'success');
        } else {
            showMessage('Ошибка входа', 'error');
        }
    })
    .catch(err => {
        showMessage('Ошибка входа: ' + err.message, 'error');
    });
}

function register(event) {
    event.preventDefault();
    const data = {
        first_name: document.getElementById('reg-first-name').value,
        last_name: document.getElementById('reg-last-name').value,
        email: document.getElementById('reg-email').value,
        phone: document.getElementById('reg-phone').value,
        password: document.getElementById('reg-password').value
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
            showMessage('Регистрация успешна', 'success');
            setTimeout(() => showPage('home'), 2000);
        } else {
            showMessage('Ошибка регистрации', 'error');
        }
    })
    .catch(err => {
        showMessage('Ошибка регистрации: ' + err.message, 'error');
    });
}

function logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('role');
    currentUser = null;
    currentRole = null;
    document.getElementById('logout-btn').style.display = 'none';
    document.getElementById('profile-link').style.display = 'none';
    document.getElementById('admin-link').style.display = 'none';
    document.getElementById('create-project-btn').style.display = 'none';
    showPage('home');
}

function loadProjects() {
    const filter = document.getElementById('project-filter').value;
    let url = `${API_BASE}/projects/projects`;
    if (filter) {
        url += `?status=${filter}`;
    }

    fetch(url)
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка загрузки проектов');
            });
        }
        return res.json();
    })
    .then(projects => {
        const list = document.getElementById('projects-list');
        if (!list) {
            console.error('projects-list element not found');
            return;
        }
        
        list.innerHTML = '';
        
        if (!projects || !Array.isArray(projects)) {
            console.error('Invalid projects data:', projects);
            showMessage('Ошибка: неверный формат данных проектов', 'error');
            return;
        }
        
        if (projects.length === 0) {
            list.innerHTML = '<p>Проекты не найдены</p>';
            return;
        }
        
        projects.forEach(project => {
            const card = document.createElement('div');
            card.className = 'project-card';
            card.innerHTML = `
                <h3>${escapeHtml(project.title || 'Без названия')}</h3>
                <p>${escapeHtml(project.short_description || '')}</p>
                <span class="status ${project.status || 'активен'}">${project.status || 'активен'}</span>
                <button onclick="viewProject('${project.id}')">Подробнее</button>
            `;
            list.appendChild(card);
        });
    })
    .catch(err => {
        showMessage('Ошибка загрузки проектов: ' + err.message, 'error');
        const list = document.getElementById('projects-list');
        if (list) {
            list.innerHTML = '<p>Не удалось загрузить проекты</p>';
        }
    });
}

function viewProject(projectId) {
    fetch(`${API_BASE}/projects/projects/${projectId}`)
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка загрузки проекта');
            });
        }
        return res.json();
    })
    .then(project => {
        document.getElementById('project-detail-title').textContent = project.title;
        const content = document.getElementById('project-detail-content');
        content.innerHTML = `
            <p><strong>Описание:</strong> ${escapeHtml(project.full_description || project.short_description || '')}</p>
            <p><strong>Статус:</strong> ${project.status}</p>
            <p><strong>Дата создания:</strong> ${new Date(project.creation_date).toLocaleDateString('ru-RU')}</p>
            <button onclick="requestParticipation('${project.id}')">Подать заявку на участие</button>
            <button onclick="showPage('projects')">Назад</button>
        `;
        showPage('project-detail');
    })
    .catch(err => {
        showMessage('Ошибка загрузки проекта: ' + err.message, 'error');
    });
}

function requestParticipation(projectId) {
    const comment = prompt('Введите комментарий к заявке:');
    if (!comment) return;

    fetch(`${API_BASE}/projects/projects/${projectId}/participate`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({ comment })
    })
    .then(res => res.json())
    .then(data => {
        showMessage('Заявка подана успешно', 'success');
    })
    .catch(err => {
        showMessage('Ошибка подачи заявки: ' + err.message, 'error');
    });
}

function loadCategories() {
    fetch(`${API_BASE}/projects/categories`)
    .then(res => res.json())
    .then(categories => {
        const select = document.getElementById('project-category');
        categories.forEach(cat => {
            const option = document.createElement('option');
            option.value = cat.id;
            option.textContent = cat.name;
            select.appendChild(option);
        });
    })
    .catch(err => console.error('Failed to load categories:', err));
}

function createProject(event) {
    event.preventDefault();
    const data = {
        title: document.getElementById('project-title').value,
        short_description: document.getElementById('project-short-desc').value,
        full_description: document.getElementById('project-full-desc').value,
        category_id: document.getElementById('project-category').value || null
    };

    fetch(`${API_BASE}/projects/projects`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify(data)
    })
    .then(res => res.json())
    .then(data => {
        if (data.id) {
            showMessage('Проект создан успешно', 'success');
            showPage('projects');
        } else {
            showMessage('Ошибка создания проекта', 'error');
        }
    })
    .catch(err => {
        showMessage('Ошибка создания проекта: ' + err.message, 'error');
    });
}

function loadProfile() {
    fetch(`${API_BASE}/users/profile`, {
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
    })
    .then(res => res.json())
    .then(user => {
        const content = document.getElementById('profile-content');
        content.innerHTML = `
            <div class="form-container">
                <p><strong>Имя:</strong> ${escapeHtml(user.first_name)}</p>
                <p><strong>Фамилия:</strong> ${escapeHtml(user.last_name)}</p>
                <p><strong>Email:</strong> ${escapeHtml(user.email)}</p>
                <p><strong>Телефон:</strong> ${escapeHtml(user.phone || 'Не указан')}</p>
                <p><strong>Роль:</strong> ${currentRole}</p>
            </div>
        `;
        if (currentRole === 'participant') {
            document.getElementById('organizer-request-section').style.display = 'block';
        }
    })
    .catch(err => {
        showMessage('Ошибка загрузки профиля: ' + err.message, 'error');
    });
}

function submitOrganizerRequest(event) {
    event.preventDefault();
    const experience = document.getElementById('organizer-experience').value;

    if (!experience || experience.trim() === '') {
        showMessage('Пожалуйста, заполните описание опыта', 'error');
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
            showMessage('Заявка подана успешно', 'success');
            document.getElementById('organizer-request-section').style.display = 'none';
            document.getElementById('organizer-experience').value = '';
        } else {
            showMessage('Ошибка подачи заявки', 'error');
        }
    })
    .catch(err => {
        showMessage('Ошибка подачи заявки: ' + err.message, 'error');
    });
}

function showAdminTab(tab) {
    document.querySelectorAll('.admin-tabs button').forEach((btn, index) => {
        btn.classList.remove('active');
        if ((tab === 'users' && index === 0) || 
            (tab === 'organizer-requests' && index === 1) || 
            (tab === 'reports' && index === 2)) {
            btn.classList.add('active');
        }
    });

    const content = document.getElementById('admin-content');
    if (tab === 'users') {
        loadUsers();
    } else if (tab === 'organizer-requests') {
        loadOrganizerRequests();
    } else if (tab === 'reports') {
        loadReports();
    }
}

function loadUsers() {
    fetch(`${API_BASE}/users/users`, {
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
    })
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка загрузки пользователей');
            });
        }
        return res.json();
    })
    .then(users => {
        const content = document.getElementById('admin-content');
        if (!content) {
            console.error('admin-content element not found');
            return;
        }
        
        if (!users || !Array.isArray(users)) {
            console.error('Invalid users data:', users);
            showMessage('Ошибка: неверный формат данных пользователей', 'error');
            content.innerHTML = '<p>Не удалось загрузить пользователей</p>';
            return;
        }
        
        if (users.length === 0) {
            content.innerHTML = '<p>Пользователи не найдены</p>';
            return;
        }
        
        content.innerHTML = `
            <table>
                <thead>
                    <tr>
                        <th>Имя</th>
                        <th>Фамилия</th>
                        <th>Email</th>
                        <th>Статус</th>
                        <th>Действия</th>
                    </tr>
                </thead>
                <tbody>
                    ${users.map(user => `
                        <tr>
                            <td>${escapeHtml(user.first_name || '')}</td>
                            <td>${escapeHtml(user.last_name || '')}</td>
                            <td>${escapeHtml(user.email || '')}</td>
                            <td>${user.status ? 'Активен' : 'Заблокирован'}</td>
                            <td>
                                <button onclick="toggleUserStatus('${user.id}', ${!user.status})">
                                    ${user.status ? 'Заблокировать' : 'Разблокировать'}
                                </button>
                            </td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        `;
    })
    .catch(err => {
        showMessage('Ошибка загрузки пользователей: ' + err.message, 'error');
        const content = document.getElementById('admin-content');
        if (content) {
            content.innerHTML = '<p>Не удалось загрузить пользователей</p>';
        }
    });
}

function toggleUserStatus(userId, status) {
    fetch(`${API_BASE}/users/users/${userId}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({ status })
    })
    .then(res => res.json())
    .then(data => {
        showMessage('Статус пользователя обновлен', 'success');
        loadUsers();
    })
    .catch(err => {
        showMessage('Ошибка обновления статуса: ' + err.message, 'error');
    });
}

function loadOrganizerRequests() {
    fetch(`${API_BASE}/users/organizer-requests`, {
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
    })
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка загрузки заявок');
            });
        }
        return res.json();
    })
    .then(requests => {
        const content = document.getElementById('admin-content');
        if (!content) {
            console.error('admin-content element not found');
            return;
        }
        
        if (!requests || !Array.isArray(requests)) {
            console.error('Invalid requests data:', requests);
            showMessage('Ошибка: неверный формат данных заявок', 'error');
            content.innerHTML = '<p>Не удалось загрузить заявки</p>';
            return;
        }
        
        if (requests.length === 0) {
            content.innerHTML = '<p>Заявки не найдены</p>';
            return;
        }
        
        content.innerHTML = `
            <table>
                <thead>
                    <tr>
                        <th>Пользователь</th>
                        <th>Описание опыта</th>
                        <th>Статус</th>
                        <th>Действия</th>
                    </tr>
                </thead>
                <tbody>
                    ${requests.map(req => `
                        <tr>
                            <td>${req.user_id || ''}</td>
                            <td>${escapeHtml(req.experience_description || '')}</td>
                            <td>${req.status || ''}</td>
                            <td>
                                ${req.status === 'в рассмотрении' ? `
                                    <button onclick="approveOrganizerRequest('${req.id}')">Одобрить</button>
                                    <button onclick="rejectOrganizerRequest('${req.id}')">Отклонить</button>
                                ` : ''}
                            </td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
        `;
    })
    .catch(err => {
        showMessage('Ошибка загрузки заявок: ' + err.message, 'error');
        const content = document.getElementById('admin-content');
        if (content) {
            content.innerHTML = '<p>Не удалось загрузить заявки</p>';
        }
    });
}

function approveOrganizerRequest(requestId) {
    fetch(`${API_BASE}/users/organizer-requests/${requestId}/approve`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
    })
    .then(res => res.json())
    .then(data => {
        showMessage('Заявка одобрена', 'success');
        loadOrganizerRequests();
    })
    .catch(err => {
        showMessage('Ошибка одобрения заявки: ' + err.message, 'error');
    });
}

function rejectOrganizerRequest(requestId) {
    const comment = prompt('Введите комментарий для отклонения:');
    fetch(`${API_BASE}/users/organizer-requests/${requestId}/reject`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({ admin_comment: comment })
    })
    .then(res => res.json())
    .then(data => {
        showMessage('Заявка отклонена', 'success');
        loadOrganizerRequests();
    })
    .catch(err => {
        showMessage('Ошибка отклонения заявки: ' + err.message, 'error');
    });
}

function loadReports() {
    fetch(`${API_BASE}/reports/reports/summary`, {
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
    })
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка загрузки отчета');
            });
        }
        return res.json();
    })
    .then(summary => {
        const content = document.getElementById('admin-content');
        if (!content) {
            console.error('admin-content element not found');
            return;
        }
        
        if (!summary) {
            console.error('Invalid summary data:', summary);
            showMessage('Ошибка: неверный формат данных отчета', 'error');
            content.innerHTML = '<p>Не удалось загрузить отчет</p>';
            return;
        }
        
        content.innerHTML = `
            <div class="form-container">
                <h3>Сводный отчет</h3>
                <p><strong>Всего пользователей:</strong> ${summary.total_users || 0}</p>
                <p><strong>Активных пользователей:</strong> ${summary.active_users || 0}</p>
                <p><strong>Всего проектов:</strong> ${summary.total_projects || 0}</p>
                <p><strong>Активных проектов:</strong> ${summary.active_projects || 0}</p>
                <p><strong>Завершенных проектов:</strong> ${summary.completed_projects || 0}</p>
                <p><strong>Всего задач:</strong> ${summary.total_tasks || 0}</p>
                <p><strong>Завершенных задач:</strong> ${summary.completed_tasks || 0}</p>
                <p><strong>Средний процент выполнения:</strong> ${(summary.average_completion_rate || 0).toFixed(2)}%</p>
            </div>
        `;
    })
    .catch(err => {
        showMessage('Ошибка загрузки отчета: ' + err.message, 'error');
        const content = document.getElementById('admin-content');
        if (content) {
            content.innerHTML = '<p>Не удалось загрузить отчет</p>';
        }
    });
}

function toggleAccessibility() {
    const panel = document.getElementById('accessibility-panel');
    panel.style.display = panel.style.display === 'none' ? 'block' : 'none';
}

function changeFontSize(size) {
    accessibilitySettings.fontSize = parseInt(size);
    document.getElementById('font-size-value').textContent = size + 'px';
    document.documentElement.style.setProperty('--font-size', size + 'px');
    saveAccessibilitySettings();
}

function toggleHighContrast(enabled) {
    accessibilitySettings.highContrast = enabled;
    if (enabled) {
        document.body.classList.add('high-contrast');
    } else {
        document.body.classList.remove('high-contrast');
    }
    saveAccessibilitySettings();
}

function toggleSimpleNav(enabled) {
    accessibilitySettings.simpleNav = enabled;
    if (enabled) {
        document.body.classList.add('simple-nav');
    } else {
        document.body.classList.remove('simple-nav');
    }
    saveAccessibilitySettings();
}

function resetAccessibility() {
    accessibilitySettings.fontSize = 16;
    accessibilitySettings.highContrast = false;
    accessibilitySettings.simpleNav = false;
    applyAccessibilitySettings();
    saveAccessibilitySettings();
    document.getElementById('font-size-slider').value = 16;
    document.getElementById('high-contrast').checked = false;
    document.getElementById('simple-nav').checked = false;
}

function applyAccessibilitySettings() {
    document.documentElement.style.setProperty('--font-size', accessibilitySettings.fontSize + 'px');
    document.getElementById('font-size-slider').value = accessibilitySettings.fontSize;
    document.getElementById('font-size-value').textContent = accessibilitySettings.fontSize + 'px';
    document.getElementById('high-contrast').checked = accessibilitySettings.highContrast;
    document.getElementById('simple-nav').checked = accessibilitySettings.simpleNav;
    
    if (accessibilitySettings.highContrast) {
        document.body.classList.add('high-contrast');
    }
    if (accessibilitySettings.simpleNav) {
        document.body.classList.add('simple-nav');
    }
}

function saveAccessibilitySettings() {
    localStorage.setItem('accessibilitySettings', JSON.stringify(accessibilitySettings));
}

function loadAccessibilitySettings() {
    const saved = localStorage.getItem('accessibilitySettings');
    if (saved) {
        Object.assign(accessibilitySettings, JSON.parse(saved));
    }
}

function showMessage(text, type) {
    const message = document.createElement('div');
    message.className = `message ${type}`;
    message.textContent = text;
    document.querySelector('main').insertBefore(message, document.querySelector('main').firstChild);
    setTimeout(() => message.remove(), 5000);
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

init();

