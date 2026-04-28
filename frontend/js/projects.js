// Модуль управления проектами

let projectTags = [];

function loadProjects() {
    const API_BASE = window.AppConfig.API_BASE;
    const filter = document.getElementById('project-filter')?.value;
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
            window.Utils.showMessage('Ошибка: неверный формат данных проектов', 'error');
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
                ? `<div class="project-tags">${project.tags.map(tag => `<span class="tag">${window.Utils.escapeHtml(tag)}</span>`).join('')}</div>`
                : '';
            const categoryHtml = project.category_name 
                ? `<div class="project-meta-item"><span>📂</span><span>${window.Utils.escapeHtml(project.category_name)}</span></div>`
                : '';
            
            card.innerHTML = `
                <h3>${window.Utils.escapeHtml(project.title || 'Без названия')}</h3>
                <p class="project-description">${window.Utils.escapeHtml(project.short_description || project.full_description || 'Описание отсутствует')}</p>
                <div class="project-meta">
                    <div class="project-meta-item">
                        <span>👤</span>
                        <span>${window.Utils.escapeHtml(project.organizer_name || 'Неизвестно')}</span>
                    </div>
                    ${categoryHtml}
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
        window.Utils.showMessage('Ошибка загрузки проектов: ' + err.message, 'error');
    });
}

function submitParticipationRequest(event) {
    event.preventDefault();
    const API_BASE = window.AppConfig.API_BASE;
    const currentProjectId = window.AppConfig.getCurrentProjectId();
    if (!currentProjectId) return;
    
    const comment = document.getElementById('participation-comment').value;
    const resume = document.getElementById('participation-resume').value;

    if (!comment || comment.trim() === '') {
        window.Utils.showMessage('Пожалуйста, заполните комментарий', 'error');
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
            window.Utils.showMessage('Заявка подана успешно', 'success');
            const participationSection = document.getElementById('participation-form-section');
            if (participationSection) {
                participationSection.innerHTML = `
                    <div class="participation-form-card">
                        <h3>Заявка отправлена</h3>
                        <p style="color: var(--text-secondary); line-height: 1.6; margin-top: var(--spacing);">
                            Ваша заявка на участие в проекте отправлена и находится на рассмотрении. 
                            Мы скоро рассмотрим вашу заявку и уведомим вас о результате.
                        </p>
                    </div>
                `;
            }
        }
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка подачи заявки: ' + err.message, 'error');
    });
}

function loadCategories() {
    const API_BASE = window.AppConfig.API_BASE;
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

function addProjectTag() {
    const input = document.getElementById('project-tag-input');
    if (!input) return;
    const tag = input.value.trim();
    if (!tag) return;
    if (projectTags.includes(tag)) {
        window.Utils.showMessage('Тег уже добавлен', 'error');
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
        <span class="tag" style="display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.25rem 0.75rem; background: var(--primary-bg); color: var(--primary); border-radius: 1rem; font-size: 0.9em;">
            ${window.Utils.escapeHtml(tag)}
            <button type="button" onclick="removeProjectTag('${window.Utils.escapeHtml(tag)}')" style="background: none; border: none; color: var(--primary); cursor: pointer; padding: 0; margin-left: 0.5rem; font-size: 1.2em; line-height: 1;">×</button>
        </span>
    `).join('');
}

function createProject(event) {
    event.preventDefault();
    const API_BASE = window.AppConfig.API_BASE;
    const title = document.getElementById('project-title').value.trim();
    const shortDesc = document.getElementById('project-short-desc').value.trim();
    const fullDesc = document.getElementById('project-full-desc').value.trim();
    const categoryId = document.getElementById('project-category').value;
    
    if (!title || !shortDesc || !fullDesc || !categoryId) {
        window.Utils.showMessage('Все поля обязательны для заполнения', 'error');
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
            // Добавляем теги после создания проекта
            if (projectTags.length > 0) {
                Promise.all(projectTags.map(tag => 
                    fetch(`${API_BASE}/projects/projects/${data.id}/tags`, {
                        method: 'POST',
                        headers: {
                            'Content-Type': 'application/json',
                            'Authorization': `Bearer ${localStorage.getItem('token')}`
                        },
                        body: JSON.stringify({ name: tag })
                    })
                )).then(() => {
                    projectTags = [];
                    updateProjectTagsDisplay();
                    window.Utils.showMessage('Проект создан', 'success');
                    window.Utils.showPage('projects');
                    loadProjects();
                }).catch(err => {
                    console.error('Error adding tags:', err);
                    projectTags = [];
                    updateProjectTagsDisplay();
                    window.Utils.showMessage('Проект создан, но не все теги добавлены', 'warning');
                    window.Utils.showPage('projects');
                    loadProjects();
                });
            } else {
                projectTags = [];
                updateProjectTagsDisplay();
                window.Utils.showMessage('Проект создан', 'success');
                window.Utils.showPage('projects');
                loadProjects();
            }
        } else {
            window.Utils.showMessage('Ошибка создания проекта', 'error');
        }
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка создания проекта: ' + err.message, 'error');
    });
}

function loadMyProjects() {
    const API_BASE = window.AppConfig.API_BASE;
    const currentRole = window.AppConfig.getCurrentRole();
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
                <h3>${window.Utils.escapeHtml(project.title)}</h3>
                <p>${window.Utils.escapeHtml(project.short_description || '')}</p>
                <p><strong>Статус:</strong> <span class="status ${project.status}">${project.status}</span></p>
                <div style="margin-top: 1rem;">
                    <button onclick="editProject('${project.id}')">Редактировать</button>
                </div>
            </div>
        `).join('');
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка загрузки проектов: ' + err.message, 'error');
    });
}

function manageProject(projectId) {
    // Предотвращаем множественные вызовы
    if (window._managingProject) {
        return;
    }
    window._managingProject = true;
    
    const API_BASE = window.AppConfig.API_BASE;
    const currentUser = window.AppConfig.getCurrentUser();
    
    // Сначала переключаемся на страницу, если нужно
    const content = document.getElementById('my-projects-content');
    if (!content) {
        window.Utils.showPage('my-projects');
        setTimeout(() => {
            window._managingProject = false;
            manageProject(projectId);
        }, 300);
        return;
    }
    
    // Показываем загрузку
    content.innerHTML = '<p style="text-align: center; padding: var(--spacing-xl);">Загрузка...</p>';
    
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
            if (err.message && err.message.includes('доступа')) throw err;
            return [];
        }),
        fetch(`${API_BASE}/projects/projects/${projectId}/participants`).then(res => {
            if (!res.ok) return [];
            return res.json();
        }).catch(() => [])
    ])
    .then(([project, requests, participants]) => {
        window._managingProject = false;
        
        if (!project || !currentUser) {
            window.Utils.showMessage('Ошибка загрузки данных проекта', 'error');
            loadMyProjects();
            return;
        }
        
        const isOrganizer = project.organizer_id === currentUser.id;
        const userParticipation = participants && Array.isArray(participants) ? participants.find(p => p.user_id === currentUser.id) : null;
        const isManager = userParticipation && (userParticipation.role === 'менеджер' || userParticipation.role === 'руководитель');
        const isLeader = userParticipation && userParticipation.role === 'руководитель';
        const canManage = isOrganizer || isManager || isLeader;
        
        if (!canManage) {
            window.Utils.showMessage('У вас нет прав для управления этим проектом', 'error');
            loadMyProjects();
            return;
        }
        
        const safeRequests = Array.isArray(requests) ? requests : [];
        const safeParticipants = Array.isArray(participants) ? participants : [];
        content.innerHTML = `
            <div class="form-container">
                <h3>Управление проектом: ${window.Utils.escapeHtml(project.title)}</h3>
                <button onclick="loadMyProjects()" class="secondary" style="margin-bottom: 1rem;">← Назад к проектам</button>
                
                <h4 style="margin-top: 2rem;">Заявки на участие</h4>
                ${safeRequests.length > 0 ? `
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
                            ${safeRequests.map(req => `
                                <tr>
                                    <td>${req.user_id || 'Неизвестно'}</td>
                                    <td>${window.Utils.escapeHtml(req.comment || '')}</td>
                                    <td>${req.resume_url ? `<a href="${window.Utils.escapeHtml(req.resume_url)}" target="_blank">Ссылка</a>` : 'Нет'}</td>
                                    <td>${req.status || 'неизвестно'}</td>
                                    <td>
                                        ${req.status === 'в рассмотрении' ? `
                                            <button onclick="window.Projects.approveRequest('${req.id}', '${projectId}')" class="success" style="margin-right: 0.5rem;">Одобрить</button>
                                            <button onclick="window.Projects.rejectRequest('${req.id}', '${projectId}')" class="danger">Отклонить</button>
                                        ` : ''}
                                    </td>
                                </tr>
                            `).join('')}
                        </tbody>
                    </table>
                ` : '<p>Заявок нет</p>'}
                
                <h4 style="margin-top: 2rem;">Участники проекта</h4>
                ${safeParticipants.length > 0 ? `
                    <table>
                        <thead>
                            <tr>
                                <th>Пользователь</th>
                                <th>Роль</th>
                                <th>Дата вступления</th>
                            </tr>
                        </thead>
                        <tbody>
                            ${safeParticipants.map(p => {
                                const login = p.email ? p.email.split('@')[0] : (p.user_id || 'Неизвестно');
                                return `
                                <tr>
                                    <td>${window.Utils.escapeHtml(login)}</td>
                                    <td>${p.role || 'участник'}</td>
                                    <td>${p.join_date ? new Date(p.join_date).toLocaleDateString('ru-RU') : 'Неизвестно'}</td>
                                </tr>
                            `;
                            }).join('')}
                        </tbody>
                    </table>
                ` : '<p>Участников нет</p>'}
            </div>
        `;
    })
    .catch(err => {
        window._managingProject = false;
        window.Utils.showMessage('Ошибка загрузки данных: ' + (err.message || 'Неизвестная ошибка'), 'error');
        console.error('Error in manageProject:', err);
        if (err.message && err.message.includes('доступа')) {
            setTimeout(() => loadMyProjects(), 1000);
        } else {
            loadMyProjects();
        }
    });
}

function approveRequest(requestId, projectId) {
    const API_BASE = window.AppConfig.API_BASE;
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
        window.Utils.showMessage('Заявка одобрена', 'success');
        manageProject(projectId);
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка одобрения заявки: ' + err.message, 'error');
    });
}

function rejectRequest(requestId, projectId) {
    const API_BASE = window.AppConfig.API_BASE;
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
        window.Utils.showMessage('Заявка отклонена', 'success');
        manageProject(projectId);
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка отклонения заявки: ' + err.message, 'error');
    });
}

function editProject(projectId) {
    const API_BASE = window.AppConfig.API_BASE;
    const currentUser = window.AppConfig.getCurrentUser();
    
    // Загружаем данные проекта
    Promise.all([
        fetch(`${API_BASE}/projects/projects/${projectId}`).then(res => {
            if (!res.ok) throw new Error('Project not found');
            return res.json();
        }),
        fetch(`${API_BASE}/projects/projects/${projectId}/tags`).then(res => res.ok ? res.json() : []).catch(() => []),
        fetch(`${API_BASE}/projects/categories`).then(res => res.ok ? res.json() : []).catch(() => []),
        fetch(`${API_BASE}/projects/projects/${projectId}/requests`, {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        }).then(res => res.ok ? res.json() : []).catch(() => []),
        fetch(`${API_BASE}/projects/projects/${projectId}/participants`).then(res => res.ok ? res.json() : []).catch(() => [])
    ])
    .then(([project, tags, categories, requests, participants]) => {
        const safeTags = Array.isArray(tags) ? tags : [];
        const safeCategories = Array.isArray(categories) ? categories : [];
        const safeRequests = Array.isArray(requests) ? requests : [];
        const safeParticipants = Array.isArray(participants) ? participants : [];
        
        const isOrganizer = currentUser && project.organizer_id === currentUser.id;
        const userParticipation = safeParticipants.find(p => p.user_id === currentUser.id);
        const isLeader = userParticipation && (userParticipation.role === 'руководитель' || userParticipation.role === 'менеджер');
        const canManage = isOrganizer || isLeader;
        
        if (!canManage) {
            window.Utils.showMessage('У вас нет прав для редактирования этого проекта', 'error');
            return;
        }
        
        const pendingRequests = safeRequests.filter(r => r.status === 'в рассмотрении');
        
        // Инициализируем массив тегов ДО создания модального окна
        _editProjectTags = [...safeTags];
        
        const modal = document.createElement('div');
        modal.className = 'task-modal-overlay';
        modal.onclick = (e) => { if (e.target === modal) modal.remove(); };
        modal.innerHTML = `
            <div class="task-modal-content" style="max-width: 95vw; max-height: 95vh; width: 1200px;">
                <div class="task-modal-header">
                    <h2 style="margin: 0; flex: 1; font-size: 20px;">Редактировать проект</h2>
                    <button class="task-modal-close" onclick="this.closest('.task-modal-overlay').remove()">&times;</button>
                </div>
                <div class="task-modal-body" style="display: grid; grid-template-columns: 1fr 1fr; gap: var(--spacing); padding: var(--spacing);">
                    <div class="task-modal-main" style="width: 100%;">
                        <h3 style="margin-bottom: var(--spacing-sm); font-size: 16px; font-weight: 600;">Информация о проекте</h3>
                        <div class="task-section" style="margin-bottom: var(--spacing-sm);">
                            <label class="task-section-label" style="font-size: 12px; margin-bottom: var(--spacing-xs);">Название проекта *</label>
                            <input type="text" id="edit-project-title" class="task-description-input" value="${window.Utils.escapeHtml(project.title || '')}" required style="padding: var(--spacing-sm); font-size: 14px;">
                        </div>
                        <div class="task-section" style="margin-bottom: var(--spacing-sm);">
                            <label class="task-section-label" style="font-size: 12px; margin-bottom: var(--spacing-xs);">Краткое описание</label>
                            <textarea id="edit-project-short-desc" class="task-description-input" rows="2" placeholder="Краткое описание проекта" style="padding: var(--spacing-sm); font-size: 14px;">${window.Utils.escapeHtml(project.short_description || '')}</textarea>
                        </div>
                        <div class="task-section" style="margin-bottom: var(--spacing-sm);">
                            <label class="task-section-label" style="font-size: 12px; margin-bottom: var(--spacing-xs);">Полное описание</label>
                            <textarea id="edit-project-full-desc" class="task-description-input" rows="4" placeholder="Полное описание проекта" style="padding: var(--spacing-sm); font-size: 14px;">${window.Utils.escapeHtml(project.full_description || '')}</textarea>
                        </div>
                        <div class="task-section" style="margin-bottom: var(--spacing-sm);">
                            <label class="task-section-label" style="font-size: 12px; margin-bottom: var(--spacing-xs);">Категория</label>
                            <select id="edit-project-category" class="task-status-select" style="padding: var(--spacing-sm); font-size: 14px;">
                                <option value="">Без категории</option>
                                ${safeCategories.map(cat => `
                                    <option value="${cat.id || ''}" ${cat.id === project.category_id ? 'selected' : ''}>${window.Utils.escapeHtml(cat.name || '')}</option>
                                `).join('')}
                            </select>
                        </div>
                        <div class="task-section" style="margin-bottom: var(--spacing-sm);">
                            <label class="task-section-label" style="font-size: 12px; margin-bottom: var(--spacing-xs);">Теги</label>
                            <div id="edit-project-tags-list" style="display: flex; flex-wrap: wrap; gap: var(--spacing-xs); margin-bottom: var(--spacing-xs);">
                                ${_editProjectTags.map((tag, index) => `
                                    <span class="tag" style="display: flex; align-items: center; gap: var(--spacing-xs); font-size: 12px; padding: 2px 8px;">
                                        ${window.Utils.escapeHtml(tag)}
                                        <button onclick="window.Projects.removeEditProjectTagByIndex(${index})" style="background: none; border: none; color: inherit; cursor: pointer; padding: 0; font-size: 12px; margin-left: 4px;">×</button>
                                    </span>
                                `).join('')}
                            </div>
                            <div style="display: flex; gap: var(--spacing-xs);">
                                <input type="text" id="edit-project-tag-input" class="task-description-input" placeholder="Добавить тег" style="flex: 1; padding: var(--spacing-sm); font-size: 14px;">
                                <button onclick="window.Projects.addEditProjectTag()" type="button" class="task-comment-submit" style="padding: var(--spacing-sm) var(--spacing); font-size: 13px;">+</button>
                            </div>
                        </div>
                    </div>
                    <div style="width: 100%; display: flex; flex-direction: column; gap: var(--spacing-sm);">
                        <div style="flex: 1; display: flex; flex-direction: column; min-height: 0;">
                            <h3 style="margin-bottom: var(--spacing-xs); font-size: 14px; font-weight: 600;">Заявки (${pendingRequests.length})</h3>
                            <div id="edit-project-requests-list" style="flex: 1; overflow-y: auto; display: flex; flex-direction: column; gap: var(--spacing-xs); min-height: 0;">
                                ${pendingRequests.length > 0 ? pendingRequests.map(req => {
                                    const login = req.email ? req.email.split('@')[0] : (req.user_id || 'Неизвестно');
                                    return `
                                        <div style="padding: var(--spacing-sm); background: var(--bg-secondary); border-radius: var(--radius-sm); border: 1px solid var(--border-light);">
                                            <div style="font-weight: 600; margin-bottom: 2px; font-size: 13px;">${window.Utils.escapeHtml(login)}</div>
                                            <div style="font-size: 12px; color: var(--text-secondary); margin-bottom: var(--spacing-xs);">${window.Utils.escapeHtml((req.comment || 'Без комментария').substring(0, 50))}${req.comment && req.comment.length > 50 ? '...' : ''}</div>
                                            ${req.resume_url ? `<div style="margin-bottom: var(--spacing-xs);"><a href="${window.Utils.escapeHtml(req.resume_url)}" target="_blank" style="color: var(--primary); font-size: 12px;">Резюме</a></div>` : ''}
                                            <div style="display: flex; gap: var(--spacing-xs);">
                                                <button onclick="approveRequestInEdit('${req.id}', '${projectId}')" class="task-comment-submit" style="flex: 1; background: var(--success); padding: var(--spacing-xs); font-size: 12px;">Одобрить</button>
                                                <button onclick="rejectRequestInEdit('${req.id}', '${projectId}')" class="task-close-btn" style="flex: 1; padding: var(--spacing-xs); font-size: 12px;">Отклонить</button>
                                            </div>
                                        </div>
                                    `;
                                }).join('') : '<div style="text-align: center; padding: var(--spacing); color: var(--text-secondary); font-size: 13px;">Нет заявок</div>'}
                            </div>
                        </div>
                        <div style="flex: 1; display: flex; flex-direction: column; min-height: 0; margin-top: var(--spacing-sm);">
                            <h3 style="margin-bottom: var(--spacing-xs); font-size: 14px; font-weight: 600;">Участники (${safeParticipants.length})</h3>
                            <div id="edit-project-participants-list" style="flex: 1; overflow-y: auto; display: flex; flex-direction: column; gap: var(--spacing-xs); min-height: 0;">
                                ${safeParticipants.length > 0 ? safeParticipants.map(p => {
                                    const login = p.email ? p.email.split('@')[0] : (p.first_name || p.user_id || 'Неизвестно');
                                    const currentRole = p.role || 'участник';
                                    const isCurrentUser = p.user_id === currentUser.id;
                                    const isOrganizerUser = project.organizer_id === p.user_id;
                                    return `
                                        <div style="padding: var(--spacing-sm); background: var(--bg-secondary); border-radius: var(--radius-sm); border: 1px solid var(--border-light);">
                                            <div style="font-weight: 600; margin-bottom: var(--spacing-xs); font-size: 13px;">${window.Utils.escapeHtml(login)}</div>
                                            <div style="display: flex; gap: var(--spacing-xs); align-items: center;">
                                                <select id="participant-role-${p.user_id}" class="task-status-select" style="flex: 1; padding: var(--spacing-xs); font-size: 13px;" ${isOrganizerUser ? 'disabled' : ''} onchange="updateParticipantRole('${p.user_id}', '${projectId}')">
                                                    <option value="участник" ${currentRole === 'участник' ? 'selected' : ''}>Участник</option>
                                                    <option value="менеджер" ${currentRole === 'менеджер' ? 'selected' : ''}>Менеджер</option>
                                                    <option value="руководитель" ${currentRole === 'руководитель' ? 'selected' : ''}>Руководитель</option>
                                                </select>
                                                ${isOrganizerUser ? '<span style="font-size: 11px; color: var(--text-secondary);">Орг.</span>' : ''}
                                            </div>
                                        </div>
                                    `;
                                }).join('') : '<div style="text-align: center; padding: var(--spacing); color: var(--text-secondary); font-size: 13px;">Нет участников</div>'}
                            </div>
                        </div>
                    </div>
                </div>
                <div class="task-sidebar-actions" style="margin: 0; padding: var(--spacing); border-top: 1px solid var(--border-light); flex-shrink: 0;">
                    <button onclick="saveProjectChanges('${projectId}')" class="task-save-btn" style="width: 100%; padding: var(--spacing-sm); font-size: 14px;">Сохранить изменения</button>
                    <button onclick="this.closest('.task-modal-overlay').remove()" class="task-close-btn" style="width: 100%; margin-top: var(--spacing-xs); padding: var(--spacing-sm); font-size: 14px;">Отмена</button>
                </div>
            </div>
        `;
        document.body.appendChild(modal);
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка загрузки данных проекта: ' + (err.message || 'Неизвестная ошибка'), 'error');
        console.error('Error loading project data:', err);
    });
}

let _editProjectTags = [];

function addEditProjectTag() {
    const input = document.getElementById('edit-project-tag-input');
    if (!input || !input.value.trim()) return;
    
    const tag = input.value.trim();
    if (_editProjectTags.includes(tag)) {
        window.Utils.showMessage('Такой тег уже добавлен', 'warning');
        return;
    }
    
    _editProjectTags.push(tag);
    input.value = '';
    updateEditProjectTagsDisplay();
}

function removeEditProjectTag(tag) {
    _editProjectTags = _editProjectTags.filter(t => t !== tag);
    updateEditProjectTagsDisplay();
}

function removeEditProjectTagByIndex(index) {
    if (index >= 0 && index < _editProjectTags.length) {
        _editProjectTags.splice(index, 1);
        updateEditProjectTagsDisplay();
    }
}

function updateEditProjectTagsDisplay() {
    const list = document.getElementById('edit-project-tags-list');
    if (!list) return;
    
    list.innerHTML = _editProjectTags.map((tag, index) => `
        <span class="tag" style="display: flex; align-items: center; gap: var(--spacing-xs); font-size: 12px; padding: 2px 8px;">
            ${window.Utils.escapeHtml(tag)}
            <button onclick="window.Projects.removeEditProjectTagByIndex(${index})" style="background: none; border: none; color: inherit; cursor: pointer; padding: 0; font-size: 12px; margin-left: 4px;">×</button>
        </span>
    `).join('');
}

function approveRequestInEdit(requestId, projectId) {
    const API_BASE = window.AppConfig.API_BASE;
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
    .then(() => {
        window.Utils.showMessage('Заявка одобрена', 'success');
        editProject(projectId);
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка одобрения заявки: ' + err.message, 'error');
    });
}

function rejectRequestInEdit(requestId, projectId) {
    const API_BASE = window.AppConfig.API_BASE;
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
    .then(() => {
        window.Utils.showMessage('Заявка отклонена', 'success');
        editProject(projectId);
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка отклонения заявки: ' + err.message, 'error');
    });
}

function updateParticipantRole(userId, projectId) {
    const API_BASE = window.AppConfig.API_BASE;
    const select = document.getElementById(`participant-role-${userId}`);
    if (!select) return;
    
    const newRole = select.value;
    
    fetch(`${API_BASE}/projects/projects/${projectId}/participants/${userId}/role`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({ role: newRole })
    })
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка обновления роли');
            });
        }
        return res.json();
    })
    .then(() => {
        window.Utils.showMessage('Роль обновлена', 'success');
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка обновления роли: ' + err.message, 'error');
        editProject(projectId);
    });
}

function saveProjectChanges(projectId) {
    const API_BASE = window.AppConfig.API_BASE;
    const title = document.getElementById('edit-project-title').value.trim();
    const shortDesc = document.getElementById('edit-project-short-desc').value.trim();
    const fullDesc = document.getElementById('edit-project-full-desc').value.trim();
    const categoryId = document.getElementById('edit-project-category').value || null;
    
    if (!title) {
        window.Utils.showMessage('Введите название проекта', 'error');
        return;
    }
    
    // Обновляем проект
    fetch(`${API_BASE}/projects/projects/${projectId}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${localStorage.getItem('token')}`
        },
        body: JSON.stringify({
            title,
            short_description: shortDesc,
            full_description: fullDesc,
            category_id: categoryId
        })
    })
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка обновления проекта');
            });
        }
        return res.json();
    })
    .then(() => {
        // Обновляем теги
        const tagsToAdd = _editProjectTags;
        return Promise.all([
            // Получаем текущие теги
            fetch(`${API_BASE}/projects/projects/${projectId}/tags`).then(res => res.ok ? res.json() : []),
            Promise.resolve(tagsToAdd)
        ]);
    })
    .then(([currentTags, newTags]) => {
        const safeCurrentTags = Array.isArray(currentTags) ? currentTags : [];
        const safeNewTags = Array.isArray(newTags) ? newTags : [];
        
        const tagsToAdd = safeNewTags.filter(t => !safeCurrentTags.includes(t));
        const tagsToRemove = safeCurrentTags.filter(t => !safeNewTags.includes(t));
        
        // Удаляем старые теги
        const deletePromises = tagsToRemove.map(tag =>
            fetch(`${API_BASE}/projects/projects/${projectId}/tags/${encodeURIComponent(tag)}`, {
                method: 'DELETE',
                headers: {
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                }
            }).catch(err => {
                console.error('Error deleting tag:', err);
                return null;
            })
        );
        
        // Добавляем новые теги
        const addPromises = tagsToAdd.map(tag =>
            fetch(`${API_BASE}/projects/projects/${projectId}/tags`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                },
                body: JSON.stringify({ name: tag })
            }).catch(err => {
                console.error('Error adding tag:', err);
                return null;
            })
        );
        
        return Promise.all([...deletePromises, ...addPromises]);
    })
    .then(() => {
        window.Utils.showMessage('Проект обновлен', 'success');
        const modal = document.querySelector('.task-modal-overlay');
        if (modal) modal.remove();
        _editProjectTags = [];
        // Перезагружаем страницу проекта
        window.viewProject(projectId);
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка обновления проекта: ' + err.message, 'error');
    });
}

window.viewProject = function(projectId) {
    const API_BASE = window.AppConfig.API_BASE;
    const currentUser = window.AppConfig.getCurrentUser();
    window.AppConfig.setCurrentProjectId(projectId);
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
        const participationSection = document.getElementById('participation-form-section');
        const sidebarInfo = document.getElementById('project-sidebar-info');
        const sidebarTitle = document.getElementById('project-sidebar-title');
        
        if (!sidebarInfo) {
            window.Utils.showMessage('Ошибка: элементы страницы не найдены', 'error');
            return;
        }
        
        const isOrganizer = currentUser && project.organizer_id === currentUser.id;
        const userParticipation = currentUser ? participants.find(p => p.user_id === currentUser.id) : null;
        const isParticipant = !!userParticipation;
        const isManager = userParticipation && (userParticipation.role === 'менеджер' || userParticipation.role === 'руководитель');
        const isLeader = userParticipation && userParticipation.role === 'руководитель';
        const canManage = isOrganizer || isLeader || isManager;
        const canRequest = currentUser && !isOrganizer && !isParticipant;
        
        Promise.all([
            fetch(`${API_BASE}/projects/projects/${projectId}/tags`).then(res => res.ok ? res.json() : []).catch(() => []),
            fetch(`${API_BASE}/projects/categories`).then(res => res.ok ? res.json() : []).catch(() => []),
            canRequest ? fetch(`${API_BASE}/projects/projects/${projectId}/my-request`, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            }).then(res => {
                if (res.status === 404) return null;
                if (res.ok) return res.json();
                return null;
            }).catch(() => null) : Promise.resolve(null)
        ])
            .then(([tags, categories, myRequest]) => {
                if (!tags || !Array.isArray(tags)) tags = [];
                const safeCategories = Array.isArray(categories) ? categories : [];
                const category = project.category_id ? safeCategories.find(c => c.id === project.category_id) : null;
                const categoryName = category ? category.name : 'Без категории';
                
                const tagsHtml = tags.length > 0 
                    ? `<div style="display: flex; flex-wrap: wrap; gap: var(--spacing-xs); margin-top: var(--spacing-xs);">${tags.map(tag => `<span class="tag">${window.Utils.escapeHtml(tag)}</span>`).join('')}</div>`
                    : '<div style="color: var(--text-tertiary); font-size: 14px; margin-top: var(--spacing-xs);">Теги не указаны</div>';
                
                if (sidebarTitle) {
                    sidebarTitle.textContent = window.Utils.escapeHtml(project.title || 'Проект');
                }
                
                // Вся информация о проекте в сайдбаре
                const description = project.full_description || project.short_description || 'Описание отсутствует';
                sidebarInfo.innerHTML = `
                    <div style="display: flex; flex-direction: column; gap: var(--spacing);">
                        <div>
                            <div style="font-size: 11px; color: var(--text-secondary); margin-bottom: var(--spacing-xs); text-transform: uppercase; letter-spacing: 0.5px; font-weight: 600;">Статус</div>
                            <div><span class="status ${project.status || 'активен'}">${project.status || 'активен'}</span></div>
                        </div>
                        <div>
                            <div style="font-size: 11px; color: var(--text-secondary); margin-bottom: var(--spacing-xs); text-transform: uppercase; letter-spacing: 0.5px; font-weight: 600;">Категория</div>
                            <div style="font-size: 14px; color: var(--text-primary); font-weight: 500;">${window.Utils.escapeHtml(categoryName)}</div>
                        </div>
                        <div>
                            <div style="font-size: 11px; color: var(--text-secondary); margin-bottom: var(--spacing-xs); text-transform: uppercase; letter-spacing: 0.5px; font-weight: 600;">Теги</div>
                            ${tagsHtml}
                        </div>
                        <div style="margin-top: var(--spacing-sm); padding-top: var(--spacing-sm); border-top: 1px solid var(--border-light);">
                            <div style="font-size: 11px; color: var(--text-secondary); margin-bottom: var(--spacing-xs); text-transform: uppercase; letter-spacing: 0.5px; font-weight: 600;">Описание</div>
                            <div style="font-size: 13px; line-height: 1.5; color: var(--text-primary); white-space: pre-wrap; max-height: 150px; overflow-y: auto;">${window.Utils.escapeHtml(description)}</div>
                        </div>
                        <div>
                            <div style="font-size: 11px; color: var(--text-secondary); margin-bottom: var(--spacing-xs); text-transform: uppercase; letter-spacing: 0.5px; font-weight: 600;">Дата создания</div>
                            <div style="font-size: 13px; color: var(--text-primary); font-weight: 500;">${new Date(project.creation_date).toLocaleDateString('ru-RU')}</div>
                        </div>
                        ${project.completion_date ? `
                            <div>
                                <div style="font-size: 11px; color: var(--text-secondary); margin-bottom: var(--spacing-xs); text-transform: uppercase; letter-spacing: 0.5px; font-weight: 600;">Дата завершения</div>
                                <div style="font-size: 13px; color: var(--text-primary); font-weight: 500;">${new Date(project.completion_date).toLocaleDateString('ru-RU')}</div>
                            </div>
                        ` : ''}
                        ${isParticipant ? `
                            <div>
                                <div style="font-size: 11px; color: var(--text-secondary); margin-bottom: var(--spacing-xs); text-transform: uppercase; letter-spacing: 0.5px; font-weight: 600;">Ваша роль</div>
                                <div style="font-size: 13px; color: var(--text-primary); font-weight: 500;">${userParticipation.role || 'участник'}</div>
                            </div>
                        ` : ''}
                        <div style="margin-top: var(--spacing); padding-top: var(--spacing); border-top: 1px solid var(--border-light);">
                            <button onclick="window.Utils.showPage('projects')" class="secondary" style="width: 100%; font-size: 13px; padding: var(--spacing-sm);">Назад к проектам</button>
                            ${canManage ? `
                                <button onclick="editProject('${project.id}')" style="width: 100%; margin-top: var(--spacing-sm); font-size: 13px; padding: var(--spacing-sm);">Редактировать проект</button>
                            ` : ''}
                            ${(isManager || isLeader) ? `
                                <button onclick="generateProjectReport('${project.id}')" style="width: 100%; margin-top: var(--spacing-sm); font-size: 13px; padding: var(--spacing-sm); background: var(--success);">Скачать отчет</button>
                            ` : ''}
                        </div>
                    </div>
                `;
                
                const participantsTab = document.getElementById('participants-tab');
                const tasksTab = document.getElementById('tasks-tab');
                const chatTab = document.getElementById('chat-tab');
                
                if (isOrganizer || isParticipant) {
                    if (participantsTab) participantsTab.style.display = 'inline-block';
                    if (tasksTab) tasksTab.style.display = 'inline-block';
                    if (chatTab) chatTab.style.display = 'inline-block';
                } else {
                    if (participantsTab) participantsTab.style.display = 'none';
                    if (tasksTab) tasksTab.style.display = 'none';
                    if (chatTab) chatTab.style.display = 'none';
                }
                
                if (canRequest) {
                    const hasPendingRequest = myRequest && myRequest.status === 'в рассмотрении';
                    if (hasPendingRequest) {
                        participationSection.innerHTML = `
                            <div class="participation-form-card">
                                <h3>Заявка отправлена</h3>
                                <p style="color: var(--text-secondary); line-height: 1.6; margin-top: var(--spacing);">
                                    Ваша заявка на участие в проекте отправлена и находится на рассмотрении. 
                                    Мы скоро рассмотрим вашу заявку и уведомим вас о результате.
                                </p>
                            </div>
                        `;
                    } else {
                        participationSection.innerHTML = `
                            <div class="participation-form-card">
                                <h3>Подать заявку на участие</h3>
                                <form onsubmit="window.Projects.submitParticipationRequest(event)">
                                    <textarea id="participation-comment" placeholder="Расскажите, почему вы хотите участвовать в этом проекте" required></textarea>
                                    <input type="url" id="participation-resume" placeholder="URL резюме (необязательно)">
                                    <button type="submit">Подать заявку</button>
                                </form>
                            </div>
                        `;
                    }
                    participationSection.style.display = 'block';
                } else {
                    participationSection.style.display = 'none';
                }
                
                if (canManage || (userParticipation && (userParticipation.role === 'менеджер' || userParticipation.role === 'руководитель'))) {
                    const createTaskBtn = document.getElementById('create-task-btn');
                    if (createTaskBtn) createTaskBtn.style.display = 'block';
                }
                
                if (isOrganizer || isParticipant) {
                    const participantsList = document.getElementById('participants-list');
                    if (participantsList) {
                        participantsList.innerHTML = participants.map(p => {
                            // Получаем логин из email (без @)
                            const login = p.email ? p.email.split('@')[0] : (p.first_name || 'Участник');
                            return `
                            <div class="participant-item">
                                <div class="participant-avatar">${(p.first_name || login || 'У')[0].toUpperCase()}</div>
                                <div class="participant-info">
                                    <div class="participant-name">${window.Utils.escapeHtml(login)}</div>
                                    <div class="participant-role">${p.role}</div>
                                </div>
                            </div>
                        `;
                        }).join('');
                    }
                }
                
                window.Utils.showPage('project-detail');
                
                if (isOrganizer || isParticipant) {
                    // По умолчанию открываем вкладку "Задачи" только для участников
                    window.showProjectDetailTab('tasks');
                } else if (canRequest) {
                    // Для не-участников показываем только форму заявки
                    document.querySelectorAll('.project-tab').forEach(tab => {
                        tab.style.display = 'none';
                    });
                    document.querySelectorAll('.project-tab-content').forEach(content => {
                        content.style.display = 'none';
                    });
                    if (participationSection) participationSection.style.display = 'block';
                }
            })
            .catch(err => console.error('Error loading tags:', err));
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка загрузки проекта: ' + err.message, 'error');
    });
};

function loadMyProjectsPage() {
    const API_BASE = window.AppConfig.API_BASE;
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
            html += '<div class="projects-grid">';
            organizedProjects.forEach(project => {
                const creationDate = new Date(project.creation_date).toLocaleDateString('ru-RU');
                html += `
                    <div class="project-card">
                        <h3>${window.Utils.escapeHtml(project.title)}</h3>
                        <p class="project-description">${window.Utils.escapeHtml(project.short_description || '')}</p>
                        <div class="project-meta">
                            <div class="project-meta-item"><span>📅</span><span>${creationDate}</span></div>
                        </div>
                        <div class="project-footer">
                            <span class="status ${project.status}">${project.status}</span>
                            <button onclick="window.viewProject('${project.id}')">Подробнее</button>
                            <button onclick="editProject('${project.id}')" class="secondary">Редактировать</button>
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
                        <h3>${window.Utils.escapeHtml(project.title)}</h3>
                        <p class="project-description">${window.Utils.escapeHtml(project.short_description || '')}</p>
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
        window.Utils.showMessage('Ошибка загрузки проектов: ' + err.message, 'error');
    });
}

function generateProjectReport(projectId) {
    const API_BASE = window.AppConfig.API_BASE;
    const token = localStorage.getItem('token');
    
    const link = document.createElement('a');
    link.href = `${API_BASE}/reports/excel/project/${projectId}`;
    link.style.display = 'none';
    document.body.appendChild(link);
    
    fetch(`${API_BASE}/reports/excel/project/${projectId}`, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(res => {
        if (!res.ok) {
            throw new Error('Ошибка генерации отчета');
        }
        return res.blob();
    })
    .then(blob => {
        const url = window.URL.createObjectURL(blob);
        link.href = url;
        link.download = `project_report_${projectId}.xlsx`;
        link.click();
        window.URL.revokeObjectURL(url);
        document.body.removeChild(link);
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка генерации отчета: ' + err.message, 'error');
        if (document.body.contains(link)) {
            document.body.removeChild(link);
        }
    });
}

// Экспорт функций
window.Projects = {
    loadProjects,
    submitParticipationRequest,
    loadCategories,
    addProjectTag,
    removeProjectTag,
    updateProjectTagsDisplay,
    createProject,
    loadMyProjects,
    manageProject,
    approveRequest,
    rejectRequest,
    loadMyProjectsPage,
    editProject,
    addEditProjectTag,
    removeEditProjectTag,
    removeEditProjectTagByIndex,
    saveProjectChanges,
    approveRequestInEdit,
    rejectRequestInEdit,
    updateParticipantRole,
    generateProjectReport
};

// Глобальные функции для использования в HTML
window.approveRequest = approveRequest;
window.rejectRequest = rejectRequest;
window.manageProject = manageProject;
window.addProjectTag = addProjectTag;
window.removeProjectTag = removeProjectTag;
window.editProject = editProject;
window.addEditProjectTag = addEditProjectTag;
window.removeEditProjectTag = removeEditProjectTag;
window.removeEditProjectTagByIndex = removeEditProjectTagByIndex;
window.saveProjectChanges = saveProjectChanges;
window.approveRequestInEdit = approveRequestInEdit;
window.rejectRequestInEdit = rejectRequestInEdit;
window.updateParticipantRole = updateParticipantRole;
window.generateProjectReport = generateProjectReport;

