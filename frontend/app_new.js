const API_BASE = '/api';

let currentUser = null;
let currentRole = null;
let currentProjectId = null;

const accessibilitySettings = {
    fontSize: 16,
    highContrast: false,
    simpleNav: false,
    largeButtons: false
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
            showPage('projects');
        })
        .catch(() => {
            localStorage.removeItem('token');
            localStorage.removeItem('role');
            showPage('home');
        });
    } else {
        showPage('home');
    }
}

function updateUI() {
    const header = document.getElementById('main-header');
    if (currentUser) {
        if (header) header.style.display = 'block';
        document.getElementById('logout-btn').style.display = 'block';
        document.getElementById('profile-link').style.display = 'block';
        if (currentRole === 'organizer' || currentRole === 'admin') {
            document.getElementById('create-project-btn').style.display = 'block';
            document.getElementById('my-projects-tab').style.display = 'block';
        }
        if (currentRole === 'admin') {
            document.getElementById('admin-link').style.display = 'block';
        }
    } else {
        if (header) header.style.display = 'none';
    }
}

function showPage(pageId) {
    if (!currentUser && pageId !== 'home' && pageId !== 'register') {
        showPage('home');
        return;
    }
    
    document.querySelectorAll('.page').forEach(page => {
        page.style.display = 'none';
    });
    
    const pageElement = document.getElementById(`${pageId}-page`);
    if (pageElement) {
        pageElement.style.display = 'block';
    }
    
    if (pageId === 'projects') {
        loadProjects();
    } else if (pageId === 'profile') {
        loadProfile();
        showProfileTab('info');
    } else if (pageId === 'admin') {
        showAdminTab('users');
    }
}

function showProfileTab(tab) {
    document.querySelectorAll('.tab-button').forEach(btn => btn.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(content => content.classList.remove('active'));
    
    if (tab === 'info') {
        document.querySelector('.tab-button[onclick="showProfileTab(\'info\')"]')?.classList.add('active');
        document.getElementById('profile-info-tab')?.classList.add('active');
    } else if (tab === 'projects') {
        document.querySelector('.tab-button[onclick="showProfileTab(\'projects\')"]')?.classList.add('active');
        document.getElementById('profile-projects-tab')?.classList.add('active');
        loadMyProjects();
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
    updateUI();
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
        if (!list) return;
        
        list.innerHTML = '';
        
        if (!projects || !Array.isArray(projects)) {
            showMessage('Ошибка: неверный формат данных проектов', 'error');
            return;
        }
        
        if (projects.length === 0) {
            list.innerHTML = '<p style="color: white; text-align: center; padding: 2rem;">Проекты не найдены</p>';
            return;
        }
        
        projects.forEach(project => {
            const card = document.createElement('div');
            card.className = 'project-card';
            card.innerHTML = `
                <h3>${escapeHtml(project.title || 'Без названия')}</h3>
                <p>${escapeHtml(project.short_description || '')}</p>
                <span class="status ${project.status || 'активен'}">${project.status || 'активен'}</span>
                <div style="margin-top: 1rem;">
                    <button onclick="viewProject('${project.id}')">Подробнее</button>
                </div>
            `;
            list.appendChild(card);
        });
    })
    .catch(err => {
        showMessage('Ошибка загрузки проектов: ' + err.message, 'error');
    });
}

function viewProject(projectId) {
    currentProjectId = projectId;
    Promise.all([
        fetch(`${API_BASE}/projects/projects/${projectId}`).then(res => res.json()),
        fetch(`${API_BASE}/projects/projects/${projectId}/participants`).then(res => res.json()).catch(() => []),
        currentUser ? fetch(`${API_BASE}/projects/projects/${projectId}/participants`, {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        }).then(res => res.json()).then(participants => {
            return participants.some(p => p.user_id === currentUser.id);
        }).catch(() => false) : Promise.resolve(false)
    ])
    .then(([project, participants, isParticipant]) => {
        document.getElementById('project-detail-title').textContent = project.title;
        const content = document.getElementById('project-detail-content');
        const participationSection = document.getElementById('participation-form-section');
        
        let canRequest = currentUser && project.organizer_id !== currentUser.id && !isParticipant;
        
        content.innerHTML = `
            <div class="form-container">
                <p><strong>Описание:</strong> ${escapeHtml(project.full_description || project.short_description || '')}</p>
                <p><strong>Статус:</strong> <span class="status ${project.status}">${project.status}</span></p>
                <p><strong>Дата создания:</strong> ${new Date(project.creation_date).toLocaleDateString('ru-RU')}</p>
                <h3 style="margin-top: 1.5rem;">Участники (${participants.length || 0})</h3>
                <div id="project-participants-list">
                    ${participants.length > 0 ? participants.map(p => `
                        <p>${p.user_id} - ${p.role}</p>
                    `).join('') : '<p>Участников пока нет</p>'}
                </div>
                <div style="margin-top: 1.5rem;">
                    <button onclick="showPage('projects')" class="secondary">Назад</button>
                </div>
            </div>
        `;
        
        if (canRequest) {
            participationSection.style.display = 'block';
        } else {
            participationSection.style.display = 'none';
        }
        
        showPage('project-detail');
    })
    .catch(err => {
        showMessage('Ошибка загрузки проекта: ' + err.message, 'error');
    });
}

function submitParticipationRequest(event) {
    event.preventDefault();
    if (!currentProjectId) return;
    
    const comment = document.getElementById('participation-comment').value;
    const resume = document.getElementById('participation-resume').value;

    if (!comment || comment.trim() === '') {
        showMessage('Пожалуйста, заполните комментарий', 'error');
        return;
    }

    fetch(`${API_BASE}/projects/projects/${currentProjectId}/participate`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({ comment, resume_url: resume || '' })
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
            document.getElementById('participation-form-section').style.display = 'none';
            document.getElementById('participation-comment').value = '';
            document.getElementById('participation-resume').value = '';
        }
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
        if (select) {
            categories.forEach(cat => {
                const option = document.createElement('option');
                option.value = cat.id;
                option.textContent = cat.name;
                select.appendChild(option);
            });
        }
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
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка создания проекта');
            });
        }
        return res.json();
    })
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
        content.innerHTML = `
            <p><strong>Имя:</strong> ${escapeHtml(user.first_name)}</p>
            <p><strong>Фамилия:</strong> ${escapeHtml(user.last_name)}</p>
            <p><strong>Email:</strong> ${escapeHtml(user.email)}</p>
            <p><strong>Телефон:</strong> ${escapeHtml(user.phone || 'Не указан')}</p>
            <p><strong>Роль:</strong> ${currentRole}</p>
        `;
        if (currentRole === 'participant') {
            document.getElementById('organizer-request-section').style.display = 'block';
        }
    })
    .catch(err => {
        showMessage('Ошибка загрузки профиля: ' + err.message, 'error');
    });
}

function loadMyProjects() {
    if (currentRole !== 'organizer' && currentRole !== 'admin') return;
    
    fetch(`${API_BASE}/projects/projects`, {
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
    })
    .then(res => res.json())
    .then(projects => {
        const myProjects = projects.filter(p => p.organizer_id === currentUser.id);
        const content = document.getElementById('my-projects-content');
        
        if (myProjects.length === 0) {
            content.innerHTML = '<p style="text-align: center; padding: 2rem; color: black;">У вас пока нет проектов</p>';
            return;
        }
        
        content.innerHTML = myProjects.map(project => `
            <div class="form-container" style="margin-bottom: 1.5rem;">
                <h3>${escapeHtml(project.title)}</h3>
                <p>${escapeHtml(project.short_description || '')}</p>
                <p><strong>Статус:</strong> <span class="status ${project.status}">${project.status}</span></p>
                <div style="margin-top: 1rem;">
                    <button onclick="manageProject('${project.id}')">Управление</button>
                </div>
            </div>
        `).join('');
    })
    .catch(err => {
        showMessage('Ошибка загрузки проектов: ' + err.message, 'error');
    });
}

function manageProject(projectId) {
    Promise.all([
        fetch(`${API_BASE}/projects/projects/${projectId}/requests`, {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        }).then(res => res.json()).catch(() => []),
        fetch(`${API_BASE}/projects/projects/${projectId}/participants`).then(res => res.json()).catch(() => [])
    ])
    .then(([requests, participants]) => {
        const content = document.getElementById('my-projects-content');
        content.innerHTML = `
            <div class="form-container">
                <h3>Управление проектом</h3>
                <button onclick="loadMyProjects()" class="secondary" style="margin-bottom: 1rem;">← Назад к проектам</button>
                
                <h4 style="margin-top: 2rem;">Заявки на участие</h4>
                ${requests.length > 0 ? `
                    <table>
                        <thead>
                            <tr>
                                <th>Пользователь</th>
                                <th>Комментарий</th>
                                <th>Статус</th>
                                <th>Действия</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${requests.map(req => `
                                <tr>
                                    <td>${req.user_id}</td>
                                    <td>${escapeHtml(req.comment || '')}</td>
                                    <td>${req.status}</td>
                                    <td>
                                        ${req.status === 'в рассмотрении' ? `
                                            <button onclick="approveRequest('${req.id}', '${projectId}')" class="success" style="margin-right: 0.5rem;">Одобрить</button>
                                            <button onclick="rejectRequest('${req.id}')" class="danger">Отклонить</button>
                                        ` : ''}
                                    </td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                ` : '<p>Заявок нет</p>'}
                
                <h4 style="margin-top: 2rem;">Участники проекта</h4>
                ${participants.length > 0 ? `
                    <table>
                        <thead>
                            <tr>
                                <th>Пользователь</th>
                                <th>Роль</th>
                                <th>Дата вступления</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${participants.map(p => `
                                <tr>
                                    <td>${p.user_id}</td>
                                    <td>${p.role}</td>
                                    <td>${new Date(p.join_date).toLocaleDateString('ru-RU')}</td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                ` : '<p>Участников нет</p>'}
            </div>
        `;
    })
    .catch(err => {
        showMessage('Ошибка загрузки данных: ' + err.message, 'error');
    });
}

function approveRequest(requestId, projectId) {
    fetch(`${API_BASE}/projects/requests/${requestId}/approve`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
    })
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка одобрения заявки');
            });
        }
        return res.json();
    })
    .then(data => {
        showMessage('Заявка одобрена', 'success');
        manageProject(projectId);
    })
    .catch(err => {
        showMessage('Ошибка одобрения заявки: ' + err.message, 'error');
    });
}

function rejectRequest(requestId) {
    if (!confirm('Вы уверены, что хотите отклонить эту заявку?')) return;
    
    fetch(`${API_BASE}/projects/requests/${requestId}/reject`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
    })
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка отклонения заявки');
            });
        }
        return res.json();
    })
    .then(data => {
        showMessage('Заявка отклонена', 'success');
        loadMyProjects();
    })
    .catch(err => {
        showMessage('Ошибка отклонения заявки: ' + err.message, 'error');
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
        if (!content) return;
        
        if (!users || !Array.isArray(users)) {
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
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка обновления статуса');
            });
        }
        return res.json();
    })
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
        if (!content) return;
        
        if (!requests || !Array.isArray(requests)) {
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
                                    <button onclick="approveOrganizerRequest('${req.id}')" class="success" style="margin-right: 0.5rem;">Одобрить</button>
                                    <button onclick="rejectOrganizerRequest('${req.id}')" class="danger">Отклонить</button>
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
    });
}

function approveOrganizerRequest(requestId) {
    fetch(`${API_BASE}/users/organizer-requests/${requestId}/approve`, {
        method: 'POST',
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
    })
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка одобрения заявки');
            });
        }
        return res.json();
    })
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
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка отклонения заявки');
            });
        }
        return res.json();
    })
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
        if (!content) return;
        
        if (!summary) {
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

function toggleLargeButtons(enabled) {
    accessibilitySettings.largeButtons = enabled;
    const buttons = document.querySelectorAll('button');
    if (enabled) {
        buttons.forEach(btn => {
            btn.style.padding = '1rem 2rem';
            btn.style.fontSize = 'calc(var(--font-size) * 1.2)';
        });
    } else {
        buttons.forEach(btn => {
            btn.style.padding = '';
            btn.style.fontSize = '';
        });
    }
    saveAccessibilitySettings();
}

function resetAccessibility() {
    accessibilitySettings.fontSize = 16;
    accessibilitySettings.highContrast = false;
    accessibilitySettings.simpleNav = false;
    accessibilitySettings.largeButtons = false;
    applyAccessibilitySettings();
    saveAccessibilitySettings();
    document.getElementById('font-size-slider').value = 16;
    document.getElementById('high-contrast').checked = false;
    document.getElementById('simple-nav').checked = false;
    document.getElementById('large-buttons').checked = false;
}

function applyAccessibilitySettings() {
    document.documentElement.style.setProperty('--font-size', accessibilitySettings.fontSize + 'px');
    const slider = document.getElementById('font-size-slider');
    const value = document.getElementById('font-size-value');
    if (slider) slider.value = accessibilitySettings.fontSize;
    if (value) value.textContent = accessibilitySettings.fontSize + 'px';
    
    const highContrast = document.getElementById('high-contrast');
    const simpleNav = document.getElementById('simple-nav');
    const largeButtons = document.getElementById('large-buttons');
    
    if (highContrast) highContrast.checked = accessibilitySettings.highContrast;
    if (simpleNav) simpleNav.checked = accessibilitySettings.simpleNav;
    if (largeButtons) largeButtons.checked = accessibilitySettings.largeButtons;
    
    if (accessibilitySettings.highContrast) {
        document.body.classList.add('high-contrast');
    }
    if (accessibilitySettings.simpleNav) {
        document.body.classList.add('simple-nav');
    }
    if (accessibilitySettings.largeButtons) {
        toggleLargeButtons(true);
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
    const main = document.querySelector('main');
    if (main) {
        main.insertBefore(message, main.firstChild);
        setTimeout(() => message.remove(), 5000);
    }
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

init();

