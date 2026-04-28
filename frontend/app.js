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
        document.getElementById('my-projects-link').style.display = 'block';
        document.getElementById('chats-link').style.display = 'block';
        // Кнопка создания проекта показывается только для организаторов
        if (currentRole === 'organizer' || currentRole === 'admin') {
            const createBtn = document.getElementById('create-project-btn');
            if (createBtn) createBtn.style.display = 'block';
            const hint = document.getElementById('create-project-hint');
            if (hint) hint.style.display = 'none';
        } else {
            const createBtn = document.getElementById('create-project-btn');
            if (createBtn) createBtn.style.display = 'none';
            const hint = document.getElementById('create-project-hint');
            if (hint) hint.style.display = 'block';
        }
        // Админ панель показывается только для админов
        if (currentRole === 'admin') {
            const adminLink = document.getElementById('admin-link');
            if (adminLink) adminLink.style.display = 'block';
        } else {
            const adminLink = document.getElementById('admin-link');
            if (adminLink) adminLink.style.display = 'none';
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
    } else if (pageId === 'admin') {
        showAdminTab('users');
    } else if (pageId === 'my-projects') {
        loadMyProjectsPage();
    } else if (pageId === 'chats') {
        loadChatsPage();
    }
}

// Функция showProfileTab удалена - профиль теперь без вкладок

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
    const firstName = document.getElementById('reg-first-name').value.trim();
    const lastName = document.getElementById('reg-last-name').value.trim();
    const email = document.getElementById('reg-email').value.trim();
    const phone = document.getElementById('reg-phone').value.trim();
    const password = document.getElementById('reg-password').value;
    
    if (!firstName || !lastName || !email || !phone || !password) {
        showMessage('Все поля обязательны для заполнения', 'error');
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
            
            const creationDate = new Date(project.creation_date).toLocaleDateString('ru-RU');
            const completionDate = project.completion_date ? new Date(project.completion_date).toLocaleDateString('ru-RU') : null;
            const tagsHtml = project.tags && project.tags.length > 0 
                ? `<div class="project-tags">${project.tags.map(tag => `<span class="tag">${escapeHtml(tag)}</span>`).join('')}</div>`
                : '';
            
            card.innerHTML = `
                <h3>${escapeHtml(project.title || 'Без названия')}</h3>
                <p class="project-description">${escapeHtml(project.short_description || project.full_description || 'Описание отсутствует')}</p>
                <div class="project-meta">
                    <div class="project-meta-item">
                        <span>👤</span>
                        <span>${escapeHtml(project.organizer_name || 'Неизвестно')}</span>
                    </div>
                    <div class="project-meta-item">
                        <span>👥</span>
                        <span>${project.participants_count || 0} участников</span>
                    </div>
                    <div class="project-meta-item">
                        <span>📅</span>
                        <span>Создан: ${creationDate}</span>
                    </div>
                    ${completionDate ? `<div class="project-meta-item"><span>✅</span><span>Завершен: ${completionDate}</span></div>` : ''}
                </div>
                ${tagsHtml}
                <div class="project-footer">
                    <span class="status ${project.status || 'активен'}">${project.status || 'активен'}</span>
                    <button onclick="window.viewProject('${project.id}')">Подробнее</button>
                </div>
            `;
            list.appendChild(card);
        });
    })
    .catch(err => {
        showMessage('Ошибка загрузки проектов: ' + err.message, 'error');
    });
}

// viewProject теперь определена в конце файла как window.viewProject

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

let projectTags = [];

function addProjectTag() {
    const input = document.getElementById('project-tag-input');
    if (!input) return;
    const tag = input.value.trim();
    if (!tag) return;
    if (projectTags.includes(tag)) {
        showMessage('Тег уже добавлен', 'error');
        return;
    }
    projectTags.push(tag);
    input.value = '';
    updateProjectTagsDisplay();
}

function removeProjectTag(tag) {
    projectTags = projectTags.filter(t => t !== tag);
    updateProjectTagsDisplay();
}

function updateProjectTagsDisplay() {
    const container = document.getElementById('project-tags-list');
    if (!container) return;
    container.innerHTML = projectTags.map(tag => `
        <span class="tag" style="display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.25rem 0.75rem; background: var(--primary-color); color: white; border-radius: 1rem; font-size: 0.9em;">
            ${escapeHtml(tag)}
            <button type="button" onclick="removeProjectTag('${escapeHtml(tag)}')" style="background: none; border: none; color: white; cursor: pointer; padding: 0; margin-left: 0.5rem; font-size: 1.2em; line-height: 1;">×</button>
        </span>
    `).join('');
}

function createProject(event) {
    event.preventDefault();
    const title = document.getElementById('project-title').value.trim();
    const shortDesc = document.getElementById('project-short-desc').value.trim();
    const fullDesc = document.getElementById('project-full-desc').value.trim();
    const categoryId = document.getElementById('project-category').value;
    
    if (!title || !shortDesc || !fullDesc || !categoryId) {
        showMessage('Все поля обязательны для заполнения', 'error');
        return;
    }
    
    const data = {
        title: title,
        short_description: shortDesc,
        full_description: fullDesc,
        category_id: categoryId || null
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
        const isOrganizer = currentRole === 'organizer' || currentRole === 'admin';
        
        content.innerHTML = `
            <div id="profile-display">
                <p><strong>Имя:</strong> ${escapeHtml(user.first_name)}</p>
                <p><strong>Фамилия:</strong> ${escapeHtml(user.last_name)}</p>
                <p><strong>Email:</strong> ${escapeHtml(user.email)}</p>
                <p><strong>Телефон:</strong> ${escapeHtml(user.phone || 'Не указан')}</p>
                <p><strong>Роль:</strong> ${currentRole}</p>
                <button onclick="window.editProfile()" style="margin-top: 1rem;">Редактировать профиль</button>
            </div>
            <div id="profile-edit" style="display:none;">
                <form onsubmit="window.saveProfile(event)">
                    <input type="text" id="edit-first-name" placeholder="Имя" value="${escapeHtml(user.first_name)}" required>
                    <input type="text" id="edit-last-name" placeholder="Фамилия" value="${escapeHtml(user.last_name)}" required>
                    <input type="tel" id="edit-phone" placeholder="Телефон" value="${escapeHtml(user.phone || '')}" required>
                    <div style="display: flex; gap: 0.5rem; margin-top: 1rem;">
                        <button type="submit">Сохранить</button>
                        <button type="button" onclick="window.cancelEditProfile()" class="secondary">Отмена</button>
                    </div>
                </form>
            </div>
        `;
        
        // Показываем заявку на роль организатора только для участников (не организаторов и не админов)
        if (currentRole === 'participant') {
            document.getElementById('organizer-request-section').style.display = 'block';
        } else {
            document.getElementById('organizer-request-section').style.display = 'none';
        }
    })
    .catch(err => {
        showMessage('Ошибка загрузки профиля: ' + err.message, 'error');
    });
}

function loadMyProjects() {
    if (currentRole !== 'organizer' && currentRole !== 'admin') return;
    
    fetch(`${API_BASE}/projects/projects/my`, {
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
    })
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка загрузки проектов');
            });
        }
        return res.json();
    })
    .then(myProjects => {
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
        fetch(`${API_BASE}/projects/projects/${projectId}`).then(res => {
            if (!res.ok) throw new Error('Project not found');
            return res.json();
        }),
        fetch(`${API_BASE}/projects/projects/${projectId}/requests`, {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        }).then(res => {
            if (!res.ok) {
                if (res.status === 403) throw new Error('Нет доступа к управлению проектом');
                return [];
            }
            return res.json();
        }).catch(err => {
            if (err.message.includes('доступа')) throw err;
            return [];
        }),
        fetch(`${API_BASE}/projects/projects/${projectId}/participants`).then(res => {
            if (!res.ok) return [];
            return res.json();
        }).catch(() => [])
    ])
    .then(([project, requests, participants]) => {
        const isOrganizer = currentUser && project.organizer_id === currentUser.id;
        const userParticipation = participants.find(p => p.user_id === currentUser.id);
        const isLeader = userParticipation && userParticipation.role === 'руководитель';
        const canManage = isOrganizer || isLeader;
        
        if (!canManage) {
            showMessage('У вас нет прав для управления этим проектом', 'error');
            loadMyProjects();
            return;
        }
        
        const content = document.getElementById('my-projects-content');
        if (!content) {
            // Если мы не на странице профиля, переключаемся на страницу "Мои проекты"
            showPage('my-projects');
            // Используем setTimeout с флагом чтобы избежать бесконечной рекурсии
            if (!window._managingProject) {
                window._managingProject = true;
                setTimeout(() => {
                    window._managingProject = false;
                    manageProject(projectId);
                }, 300);
            }
            return;
        }
        content.innerHTML = `
            <div class="form-container">
                <h3>Управление проектом: ${escapeHtml(project.title)}</h3>
                <button onclick="loadMyProjects()" class="secondary" style="margin-bottom: 1rem;">← Назад к проектам</button>
                
                <h4 style="margin-top: 2rem;">Заявки на участие</h4>
                ${requests.length > 0 ? `
                    <table>
                        <thead>
                            <tr>
                                <th>Пользователь</th>
                                <th>Комментарий</th>
                                <th>Резюме</th>
                                <th>Статус</th>
                                <th>Действия</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${requests.map(req => `
                                <tr>
                                    <td>${req.user_id}</td>
                                    <td>${escapeHtml(req.comment || '')}</td>
                                    <td>${req.resume_url ? `<a href="${escapeHtml(req.resume_url)}" target="_blank">Ссылка</a>` : 'Нет'}</td>
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
        if (err.message.includes('доступа')) {
            setTimeout(() => loadMyProjects(), 1000);
        }
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
            // Проверяем, является ли ответ JSON
            const contentType = res.headers.get('content-type');
            if (contentType && contentType.includes('application/json')) {
                return res.json().then(data => {
                    throw new Error(data.error || 'Ошибка загрузки отчета');
                });
            } else {
                // Если не JSON, читаем как текст
                return res.text().then(text => {
                    throw new Error(text || 'Ошибка загрузки отчета');
                });
            }
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
    if (enabled) {
        document.body.classList.add('large-buttons');
    } else {
        document.body.classList.remove('large-buttons');
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
    toggleLargeButtons(accessibilitySettings.largeButtons);
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

let currentChatWS = null;
let currentChatId = null;

function loadProjectTasks(projectId) {
    const tasksSection = document.getElementById('project-tasks-section');
    const tasksContainer = document.getElementById('tasks-container');
    if (!tasksSection || !tasksContainer) return;
    tasksSection.style.display = 'block';
    tasksContainer.innerHTML = '<p>Загрузка задач...</p>';
    fetch(`${API_BASE}/tasks/projects/${projectId}/tasks`, {
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
    })
    .then(res => res.ok ? res.json() : [])
    .then(tasks => {
        if (!tasks || !Array.isArray(tasks) || tasks.length === 0) {
            tasksContainer.innerHTML = '<p>Задач пока нет</p>';
            return;
        }
        // Notion-стиль: колонки по статусам
        const statusColumns = {
            'новая': [],
            'в работе': [],
            'на проверке': [],
            'тестирование': [],
            'завершена': [],
            'архив': []
        };
        
        tasks.forEach(task => {
            if (statusColumns[task.status]) {
                statusColumns[task.status].push(task);
            } else {
                statusColumns['новая'].push(task);
            }
        });
        
        tasksContainer.innerHTML = `
            <div class="notion-board">
                ${Object.entries(statusColumns).map(([status, statusTasks]) => `
                    <div class="notion-column">
                        <div class="notion-column-header">
                            <h4>${status}</h4>
                            <span class="task-count">${statusTasks.length}</span>
                        </div>
                        <div class="notion-column-content">
                            ${statusTasks.map(task => `
                                <div class="notion-task-card" draggable="true" data-task-id="${task.id}" data-status="${task.status}">
                                    <div class="task-card-header">
                                        <h5>${escapeHtml(task.title)}</h5>
                                    </div>
                                    <p class="task-card-description">${escapeHtml((task.description || '').substring(0, 100))}${task.description && task.description.length > 100 ? '...' : ''}</p>
                                    <div class="task-card-footer">
                                        <span class="task-date-small">${new Date(task.creation_date).toLocaleDateString('ru-RU')}</span>
                                        ${task.assigned_to ? `<span class="task-assigned-badge">👤</span>` : ''}
                                    </div>
                                    <div class="task-card-actions">
                                        <button onclick="viewTask('${task.id}')" class="task-card-btn">Открыть</button>
                                    </div>
                                </div>
                            `).join('')}
                            ${statusTasks.length === 0 ? '<div class="empty-column">Нет задач</div>' : ''}
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    })
    .catch(err => { tasksContainer.innerHTML = '<p>Ошибка загрузки задач: ' + err.message + '</p>'; });
}

function showCreateTaskForm() {
    const title = prompt('Название задачи:');
    if (!title) return;
    const description = prompt('Описание задачи:');
    fetch(`${API_BASE}/tasks/projects/${currentProjectId}/tasks`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${localStorage.getItem('token')}` },
        body: JSON.stringify({ title, description: description || '' })
    })
    .then(res => { if (!res.ok) return res.json().then(data => { throw new Error(data.error || 'Ошибка создания задачи'); }); return res.json(); })
    .then(() => { showMessage('Задача создана', 'success'); loadProjectTasks(currentProjectId); })
    .catch(err => { showMessage('Ошибка создания задачи: ' + err.message, 'error'); });
}

function viewTask(taskId) {
    fetch(`${API_BASE}/tasks/tasks/${taskId}`, { headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` } })
    .then(res => { if (!res.ok) throw new Error('Task not found'); return res.json(); })
    .then(task => {
        Promise.all([
            fetch(`${API_BASE}/tasks/tasks/${taskId}/comments`).then(res => res.ok ? res.json() : []),
            fetch(`${API_BASE}/projects/projects/${task.project_id}`).then(res => res.ok ? res.json() : null)
        ])
        .then(([comments, project]) => {
            const modal = document.createElement('div');
            modal.className = 'modal';
            modal.innerHTML = `<div class="modal-content">
                <span class="close" onclick="this.closest('.modal').remove()">&times;</span>
                <h2>${escapeHtml(task.title)}</h2>
                <p><strong>Описание:</strong> ${escapeHtml(task.description || '')}</p>
                <p><strong>Статус:</strong> <span class="status ${task.status}">${task.status}</span></p>
                <p><strong>Дата создания:</strong> ${new Date(task.creation_date).toLocaleDateString('ru-RU')}</p>
                ${task.assigned_to ? `<p><strong>Назначено:</strong> ${task.assigned_to}</p>` : ''}
                <h3>Комментарии</h3>
                <div id="task-comments">${comments.map(c => `<div class="comment"><p>${escapeHtml(c.content)}</p>
                <small>${new Date(c.publication_date).toLocaleDateString('ru-RU')}</small></div>`).join('')}</div>
                <form onsubmit="addTaskComment(event, '${taskId}')">
                    <textarea id="task-comment-input" placeholder="Добавить комментарий" required></textarea>
                    <button type="submit">Добавить комментарий</button>
                </form>
                <div style="margin-top: 1rem;">
                    <button onclick="updateTaskStatus('${taskId}')" class="secondary">Изменить статус</button>
                    <button onclick="this.closest('.modal').remove()" class="secondary">Закрыть</button>
                </div>
            </div>`;
            document.body.appendChild(modal);
        });
    })
    .catch(err => { showMessage('Ошибка загрузки задачи: ' + err.message, 'error'); });
}

function addTaskComment(event, taskId) {
    event.preventDefault();
    const content = document.getElementById('task-comment-input').value;
    if (!content) return;
    fetch(`${API_BASE}/tasks/tasks/${taskId}/comments`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${localStorage.getItem('token')}` },
        body: JSON.stringify({ content })
    })
    .then(res => { if (!res.ok) throw new Error('Failed to add comment'); showMessage('Комментарий добавлен', 'success'); viewTask(taskId); })
    .catch(err => { showMessage('Ошибка добавления комментария: ' + err.message, 'error'); });
}

function updateTaskStatus(taskId) {
    const statuses = ['новая', 'в работе', 'на проверке', 'тестирование', 'завершена', 'архив'];
    const currentStatus = prompt('Выберите статус:\n' + statuses.map((s, i) => `${i + 1}. ${s}`).join('\n'));
    if (!currentStatus) return;
    const statusIndex = parseInt(currentStatus) - 1;
    if (statusIndex < 0 || statusIndex >= statuses.length) { showMessage('Неверный статус', 'error'); return; }
    fetch(`${API_BASE}/tasks/tasks/${taskId}/status`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${localStorage.getItem('token')}` },
        body: JSON.stringify({ status: statuses[statusIndex] })
    })
    .then(res => { if (!res.ok) throw new Error('Failed to update status'); showMessage('Статус обновлен', 'success'); loadProjectTasks(currentProjectId); })
    .catch(err => { showMessage('Ошибка обновления статуса: ' + err.message, 'error'); });
}

function loadProjectChat(projectId) {
    const chatSection = document.getElementById('project-chat-section');
    const chatMessages = document.getElementById('chat-messages');
    if (!chatSection || !chatMessages) return;
    chatSection.style.display = 'block';
    chatMessages.innerHTML = '<p>Загрузка чата...</p>';
    fetch(`${API_BASE}/chats/projects/${projectId}/chats`, { headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` } })
    .then(res => res.ok ? res.json() : [])
    .then(chats => {
        if (!chats || chats.length === 0) { createProjectChat(projectId); return; }
        currentChatId = chats[0].id;
        loadChatMessages(chats[0].id);
        connectWebSocket(chats[0].id);
    })
    .catch(err => { console.error('Error loading chat:', err); chatMessages.innerHTML = '<p>Ошибка загрузки чата</p>'; });
}

function createProjectChat(projectId) {
    fetch(`${API_BASE}/chats/projects/${projectId}/chats`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${localStorage.getItem('token')}` },
        body: JSON.stringify({ name: 'Основной чат' })
    })
    .then(res => { 
        if (!res.ok) {
            return res.text().then(text => {
                throw new Error(text || 'Failed to create chat');
            });
        }
        return res.json(); 
    })
    .then(data => { 
        currentChatId = data.id; 
        currentProjectId = projectId;
        
        // Обновляем интерфейс чата
        const chatSection = document.getElementById('project-chat-section');
        const content = document.getElementById('chats-content');
        
        if (chatSection) {
            // Если мы на странице проекта
            chatSection.style.display = 'block';
            const chatMessages = document.getElementById('chat-messages');
            if (chatMessages) {
                setTimeout(() => {
                    loadChatMessages(data.id);
                    connectWebSocket(data.id);
                }, 100);
            }
        } else if (content) {
            // Если мы на странице чатов
            content.innerHTML = `
                <div class="form-container">
                    <h3>Чат проекта</h3>
                    <div id="chat-messages" style="max-height: 500px; overflow-y: auto; margin-bottom: 1rem; padding: 1rem; background: var(--background-color); border-radius: var(--radius); min-height: 300px;"></div>
                    <div id="chat-input-container" style="display: flex; gap: 0.5rem;">
                        <input type="text" id="chat-message-input" placeholder="Введите сообщение..." style="flex: 1;" onkeypress="if(event.key==='Enter') sendChatMessage()">
                        <button onclick="sendChatMessage()">Отправить</button>
                    </div>
                </div>
            `;
            // Загружаем сообщения после создания элементов
            setTimeout(() => {
                loadChatMessages(data.id);
                connectWebSocket(data.id);
            }, 100);
        }
    })
    .catch(err => { 
        console.error('Error creating chat:', err);
        showMessage('Ошибка создания чата: ' + err.message, 'error');
        const content = document.getElementById('chats-content');
        if (content) content.innerHTML = '<p>Ошибка создания чата</p>';
    });
}

function loadChatMessages(chatId) {
    if (!chatId) return;
    fetch(`${API_BASE}/chats/chats/${chatId}/messages`, { 
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` } 
    })
    .then(res => {
        if (!res.ok) return [];
        return res.json();
    })
    .then(messages => {
        const chatMessages = document.getElementById('chat-messages');
        if (!chatMessages) return;
        if (!messages || !Array.isArray(messages)) messages = [];
        if (messages.length === 0) {
            chatMessages.innerHTML = '<p style="text-align: center; color: #999; padding: 2rem;">Нет сообщений</p>';
            return;
        }
        chatMessages.innerHTML = messages.map(msg => `
            <div class="chat-message ${msg.user_id === currentUser.id ? 'own' : ''}">
                <div class="message-header"><strong>${msg.user_id === currentUser.id ? 'Вы' : 'Участник'}</strong>
                <span class="message-time">${new Date(msg.sent_at).toLocaleString('ru-RU')}</span></div>
                <div class="message-text">${escapeHtml(msg.message_text)}</div>
            </div>
        `).join('');
        chatMessages.scrollTop = chatMessages.scrollHeight;
    })
    .catch(err => { 
        console.error('Error loading messages:', err);
        const chatMessages = document.getElementById('chat-messages');
        if (chatMessages) chatMessages.innerHTML = '<p style="color: red;">Ошибка загрузки сообщений</p>';
    });
}

function connectWebSocket(chatId) {
    if (currentChatWS) currentChatWS.close();
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/chats/chats/${chatId}/ws`;
    currentChatWS = new WebSocket(wsUrl);
    currentChatWS.onopen = () => { console.log('WebSocket connected'); };
    currentChatWS.onmessage = (event) => { const message = JSON.parse(event.data); addMessageToChat(message); };
    currentChatWS.onerror = (error) => { console.error('WebSocket error:', error); };
    currentChatWS.onclose = () => { console.log('WebSocket disconnected'); setTimeout(() => { if (currentChatId) connectWebSocket(currentChatId); }, 3000); };
}

function addMessageToChat(message) {
    const chatMessages = document.getElementById('chat-messages');
    if (!chatMessages) return;
    const messageDiv = document.createElement('div');
    messageDiv.className = `chat-message ${message.user_id === currentUser.id ? 'own' : ''}`;
    messageDiv.innerHTML = `<div class="message-header"><strong>${message.user_id === currentUser.id ? 'Вы' : 'Участник'}</strong>
    <span class="message-time">${new Date(message.sent_at).toLocaleString('ru-RU')}</span></div>
    <div class="message-text">${escapeHtml(message.message_text)}</div>`;
    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
}

function sendChatMessage() {
    if (!currentChatId || !currentUser) {
        showMessage('Чат не загружен', 'error');
        return;
    }
    
    const input = document.getElementById('chat-message-input');
    if (!input) {
        showMessage('Поле ввода не найдено', 'error');
        return;
    }
    
    const messageText = input.value;
    if (!messageText || messageText.trim().length === 0) {
        showMessage('Введите сообщение', 'error');
        return;
    }
    
    const trimmedText = messageText.trim();
    
    if (currentChatWS && currentChatWS.readyState === WebSocket.OPEN) {
        currentChatWS.send(JSON.stringify({ user_id: currentUser.id, message_text: trimmedText }));
        input.value = '';
    } else {
        // Fallback на HTTP если WebSocket не работает
        fetch(`${API_BASE}/chats/chats/${currentChatId}/messages`, {
            method: 'POST',
            headers: { 
                'Content-Type': 'application/json', 
                'Authorization': `Bearer ${localStorage.getItem('token')}` 
            },
            body: JSON.stringify({ message_text: trimmedText })
        })
        .then(res => {
            if (!res.ok) {
                return res.text().then(text => {
                    throw new Error(text || 'Ошибка отправки сообщения');
                });
            }
            return res.json();
        })
        .then(() => {
            input.value = '';
            loadChatMessages(currentChatId);
        })
        .catch(err => {
            showMessage('Ошибка отправки сообщения: ' + err.message, 'error');
        });
    }
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
    const firstName = document.getElementById('edit-first-name').value;
    const lastName = document.getElementById('edit-last-name').value;
    const phone = document.getElementById('edit-phone').value;
    
    if (!firstName || !lastName || !phone) {
        showMessage('Все поля обязательны для заполнения', 'error');
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
        showMessage('Профиль обновлен', 'success');
        loadProfile();
    })
    .catch(err => {
        showMessage('Ошибка обновления профиля: ' + err.message, 'error');
    });
};

// Исправляем viewProject чтобы она была доступна глобально
window.viewProject = function(projectId) {
    currentProjectId = projectId;
    Promise.all([
        fetch(`${API_BASE}/projects/projects/${projectId}`).then(res => {
            if (!res.ok) throw new Error('Project not found');
            return res.json();
        }),
        fetch(`${API_BASE}/projects/projects/${projectId}/participants`).then(res => {
            if (res.ok) return res.json();
            return [];
        }).catch(() => []),
        currentUser ? fetch(`${API_BASE}/projects/projects/my`, {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        }).then(res => {
            if (res.ok) return res.json();
            return [];
        }).catch(() => []) : Promise.resolve([])
    ])
    .then(([project, participants, myProjects]) => {
        const titleEl = document.getElementById('project-detail-title');
        const content = document.getElementById('project-detail-content');
        const participationSection = document.getElementById('participation-form-section');
        
        if (!titleEl || !content) {
            showMessage('Ошибка: элементы страницы не найдены', 'error');
            return;
        }
        
        titleEl.textContent = project.title;
        
        const isOrganizer = currentUser && project.organizer_id === currentUser.id;
        const userParticipation = currentUser ? participants.find(p => p.user_id === currentUser.id) : null;
        const isParticipant = !!userParticipation;
        const isLeader = userParticipation && userParticipation.role === 'руководитель';
        const canManage = isOrganizer || isLeader;
        const canRequest = currentUser && !isOrganizer && !isParticipant;
        
        fetch(`${API_BASE}/projects/projects/${projectId}/tags`)
            .then(res => res.ok ? res.json() : [])
            .then(tags => {
                if (!tags || !Array.isArray(tags)) tags = [];
                const tagsHtml = tags.length > 0 
                    ? `<div class="project-tags" style="margin: 1rem 0;">${tags.map(tag => `<span class="tag">${escapeHtml(tag)}</span>`).join('')}</div>`
                    : '';
                
                content.innerHTML = `
                    <div class="form-container">
                        <p><strong>Описание:</strong> ${escapeHtml(project.full_description || project.short_description || '')}</p>
                        <p><strong>Статус:</strong> <span class="status ${project.status}">${project.status}</span></p>
                        <p><strong>Дата создания:</strong> ${new Date(project.creation_date).toLocaleDateString('ru-RU')}</p>
                        ${project.completion_date ? `<p><strong>Дата завершения:</strong> ${new Date(project.completion_date).toLocaleDateString('ru-RU')}</p>` : ''}
                        ${isParticipant ? `<p><strong>Ваша роль:</strong> ${userParticipation.role}</p>` : ''}
                        ${tagsHtml}
                        <div style="margin-top: 1.5rem;">
                            <button onclick="showPage('projects')" class="secondary">Назад</button>
                            ${canManage ? `<button onclick="manageProject('${project.id}')" style="margin-left: 0.5rem;">Управление</button>` : ''}
                            ${(isOrganizer || isParticipant) ? `<button onclick="showProjectTasks('${project.id}')" style="margin-left: 0.5rem;">Задачи</button>` : ''}
                            ${(isOrganizer || isParticipant) ? `<button onclick="loadProjectChat('${project.id}')" style="margin-left: 0.5rem;">Чат</button>` : ''}
                        </div>
                    </div>
                `;
                
                if (canRequest) {
                    participationSection.style.display = 'block';
                } else {
                    participationSection.style.display = 'none';
                    if (isOrganizer) {
                        content.innerHTML += '<p style="color: var(--info-color); margin-top: 1rem;">Вы являетесь организатором этого проекта</p>';
                    } else if (isParticipant) {
                        content.innerHTML += `<p style="color: var(--success-color); margin-top: 1rem;">Вы уже являетесь участником этого проекта (${userParticipation.role})</p>`;
                    }
                }
                
                if (canManage) {
                    const createTaskBtn = document.getElementById('create-task-btn');
                    if (createTaskBtn) createTaskBtn.style.display = 'block';
                }
                
                showPage('project-detail');
            })
            .catch(err => console.error('Error loading tags:', err));
    })
    .catch(err => {
        showMessage('Ошибка загрузки проекта: ' + err.message, 'error');
    });
};

function loadChatsPage() {
    const selector = document.getElementById('chat-project-select');
    const content = document.getElementById('chats-content');
    
    if (!selector || !content) return;
    
    content.innerHTML = '<p>Загрузка проектов...</p>';
    
    // Загружаем проекты пользователя (где он участвует)
    fetch(`${API_BASE}/projects/projects/my`, {
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
    })
    .then(res => {
        if (!res.ok) return [];
        return res.json();
    })
    .then(projects => {
        selector.innerHTML = '<option value="">Выберите проект</option>';
        if (projects && Array.isArray(projects) && projects.length > 0) {
            projects.forEach(project => {
                const option = document.createElement('option');
                option.value = project.id;
                option.textContent = project.title;
                selector.appendChild(option);
            });
        } else {
            selector.innerHTML = '<option value="">У вас нет проектов</option>';
        }
    })
    .catch(err => {
        console.error('Error loading projects:', err);
        selector.innerHTML = '<option value="">Ошибка загрузки</option>';
    });
    
    content.innerHTML = '<p style="text-align: center; color: white; padding: 2rem;">Выберите проект для просмотра чата</p>';
}

function loadProjectChats() {
    const projectId = document.getElementById('chat-project-select').value;
    const content = document.getElementById('chats-content');
    
    if (!projectId) {
        if (content) content.innerHTML = '<p style="text-align: center; color: white; padding: 2rem;">Выберите проект</p>';
        return;
    }
    
    if (content) content.innerHTML = '<p>Загрузка чата...</p>';
    currentProjectId = projectId;
    
    fetch(`${API_BASE}/chats/projects/${projectId}/chats`, {
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
    })
    .then(res => {
        if (!res.ok) return [];
        return res.json();
    })
    .then(chats => {
        if (!chats || chats.length === 0) {
            // Создаем чат если его нет
            createProjectChat(projectId);
            return;
        }
        
        const chat = chats[0];
        currentChatId = chat.id;
        
        if (content) {
            content.innerHTML = `
                <div class="form-container">
                    <h3>Чат проекта</h3>
                    <div id="chat-messages" style="max-height: 500px; overflow-y: auto; margin-bottom: 1rem; padding: 1rem; background: var(--background-color); border-radius: var(--radius); min-height: 300px;"></div>
                    <div id="chat-input-container" style="display: flex; gap: 0.5rem;">
                        <input type="text" id="chat-message-input" placeholder="Введите сообщение..." style="flex: 1;" onkeypress="if(event.key==='Enter') { event.preventDefault(); sendChatMessage(); }">
                        <button onclick="sendChatMessage()">Отправить</button>
                    </div>
                </div>
            `;
            
            // Загружаем сообщения после создания элементов
            setTimeout(() => {
                loadChatMessages(chat.id);
                connectWebSocket(chat.id);
            }, 100);
        }
    })
    .catch(err => {
        console.error('Error loading chat:', err);
        if (content) content.innerHTML = '<p>Ошибка загрузки чата</p>';
    });
}

window.showProjectTasks = function(projectId) {
    currentProjectId = projectId;
    const tasksSection = document.getElementById('project-tasks-section');
    if (tasksSection) {
        tasksSection.style.display = 'block';
        loadProjectTasks(projectId);
    } else {
        window.viewProject(projectId);
        setTimeout(() => {
            const tasksSection = document.getElementById('project-tasks-section');
            if (tasksSection) {
                tasksSection.style.display = 'block';
                loadProjectTasks(projectId);
            }
        }, 500);
    }
};

function loadMyProjectsPage() {
    const list = document.getElementById('my-projects-list');
    if (!list) return;
    
    list.innerHTML = '<p>Загрузка...</p>';
    
    fetch(`${API_BASE}/projects/projects/my`, {
        headers: {
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        }
    })
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка загрузки проектов');
            });
        }
        return res.json();
    })
    .then(myProjects => {
        if (!myProjects || !Array.isArray(myProjects) || myProjects.length === 0) {
            list.innerHTML = '<p style="text-align: center; padding: 2rem; color: black;">У вас пока нет проектов</p>';
            return;
        }
        
        const organizedProjects = myProjects.filter(p => p.user_role === 'organizer');
        const participatedProjects = myProjects.filter(p => p.user_role !== 'organizer');
        
        let html = '';
        
        if (organizedProjects.length > 0) {
            html += '<h3 style="color: white; margin-bottom: 1rem;">Мои проекты (организатор)</h3>';
            html += '<div class="projects-grid">';
            organizedProjects.forEach(project => {
                const creationDate = new Date(project.creation_date).toLocaleDateString('ru-RU');
                html += `
                    <div class="project-card">
                        <h3>${escapeHtml(project.title)}</h3>
                        <p class="project-description">${escapeHtml(project.short_description || '')}</p>
                        <div class="project-meta">
                            <div class="project-meta-item"><span>👥</span><span>${project.participants_count || 0} участников</span></div>
                            <div class="project-meta-item"><span>📅</span><span>${creationDate}</span></div>
                        </div>
                        <div class="project-footer">
                            <span class="status ${project.status}">${project.status}</span>
                            <button onclick="window.viewProject('${project.id}')">Подробнее</button>
                            <button onclick="manageProject('${project.id}')" class="secondary">Управление</button>
                        </div>
                    </div>
                `;
            });
            html += '</div>';
        }
        
        if (participatedProjects.length > 0) {
            html += '<h3 style="color: black; margin-top: 2rem; margin-bottom: 1rem;">Проекты, где я участвую</h3>';
            html += '<div class="projects-grid">';
            participatedProjects.forEach(project => {
                const creationDate = new Date(project.creation_date).toLocaleDateString('ru-RU');
                html += `
                    <div class="project-card">
                        <h3>${escapeHtml(project.title)}</h3>
                        <p class="project-description">${escapeHtml(project.short_description || '')}</p>
                        <div class="project-meta">
                            <div class="project-meta-item"><span>📅</span><span>${creationDate}</span></div>
                            <div class="project-meta-item"><span>Роль:</span><span>${project.user_role === 'leader' ? 'Руководитель' : 'Участник'}</span></div>
                        </div>
                        <div class="project-footer">
                            <span class="status ${project.status}">${project.status}</span>
                            <button onclick="window.viewProject('${project.id}')">Подробнее</button>
                        </div>
                    </div>
                `;
            });
            html += '</div>';
        }
        
        list.innerHTML = html || '<p style="text-align: center; padding: 2rem; color: black;">У вас пока нет проектов</p>';
    })
    .catch(err => {
        showMessage('Ошибка загрузки проектов: ' + err.message, 'error');
    });
}

init();

