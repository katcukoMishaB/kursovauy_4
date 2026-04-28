// Модуль управления задачами

function loadProjectTasks(projectId) {
    const API_BASE = window.AppConfig.API_BASE;
    const tasksContainer = document.getElementById('tasks-container');
    if (!tasksContainer) return;
    
    tasksContainer.innerHTML = '<p style="text-align: center; padding: var(--spacing-xl); color: var(--text-secondary);">Загрузка задач...</p>';
    
    Promise.all([
        fetch(`${API_BASE}/tasks/projects/${projectId}/tasks`, {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        }).then(res => res.ok ? res.json() : []),
        fetch(`${API_BASE}/projects/projects/${projectId}/participants`).then(res => res.ok ? res.json() : [])
    ])
    .then(([tasks, participants]) => {
        if (!tasks || !Array.isArray(tasks) || tasks.length === 0) {
            tasksContainer.innerHTML = '<div class="empty-column" style="text-align: center; padding: var(--spacing-xl); color: var(--text-secondary);">Задач пока нет</div>';
            return;
        }
        
        // Канбан-доска: колонки по статусам
        const statusColumns = {
            'новая': [],
            'в работе': [],
            'на проверке': [],
            'завершена': []
        };
        
        tasks.forEach(task => {
            const status = task.status || 'новая';
            if (statusColumns[status]) {
                statusColumns[status].push(task);
            } else {
                statusColumns['новая'].push(task);
            }
        });
        
        // Создаем мапу участников для быстрого доступа
        const participantsMap = {};
        participants.forEach(p => {
            participantsMap[p.user_id] = p;
        });
        
        tasksContainer.innerHTML = `
            <div class="kanban-board" id="kanban-board">
                ${Object.entries(statusColumns).map(([status, statusTasks]) => `
                    <div class="kanban-column" data-status="${status}" ondrop="handleDrop(event)" ondragover="handleDragOver(event)">
                        <div class="kanban-column-header">
                            <h4>${status}</h4>
                            <span class="task-count">${statusTasks.length}</span>
                        </div>
                        <div class="kanban-column-content" id="column-${status}">
                            ${statusTasks.map(task => {
                                const assignedUser = task.assigned_to ? participantsMap[task.assigned_to] : null;
                                const assigneeName = assignedUser ? (assignedUser.email ? assignedUser.email.split('@')[0] : `User ${task.assigned_to}`) : null;
                                return `
                                <div class="kanban-task-card" draggable="true" data-task-id="${task.id}" data-status="${task.status}" ondragstart="handleDragStart(event)" onclick="viewTask('${task.id}')">
                                    <div class="task-card-id">#${task.id}</div>
                                    <div class="task-card-title">${window.Utils.escapeHtml(task.title)}</div>
                                    <div class="task-card-footer">
                                        ${assigneeName ? `<div class="task-assignee">👤 ${window.Utils.escapeHtml(assigneeName)}</div>` : '<div class="task-assignee-empty">Не назначена</div>'}
                                        ${task.comments_count ? `<div class="task-comments-count">💬 ${task.comments_count}</div>` : ''}
                                    </div>
                                </div>
                            `;
                            }).join('')}
                            ${statusTasks.length === 0 ? '<div class="empty-column">Нет задач</div>' : ''}
                        </div>
                    </div>
                `).join('')}
            </div>
        `;
    })
    .catch(err => { 
        tasksContainer.innerHTML = '<p style="color: var(--danger); text-align: center; padding: var(--spacing-xl);">Ошибка загрузки задач: ' + err.message + '</p>'; 
    });
}

function showCreateTaskForm() {
    const currentProjectId = window.AppConfig.getCurrentProjectId();
    if (!currentProjectId) {
        window.Utils.showMessage('Проект не выбран', 'error');
        return;
    }
    
    const modal = document.createElement('div');
    modal.className = 'task-modal-overlay';
    modal.onclick = (e) => { if (e.target === modal) modal.remove(); };
    modal.innerHTML = `
        <div class="task-modal-content" style="max-width: 600px;">
            <div class="task-modal-header">
                <h2 style="margin: 0; flex: 1;">Создать задачу</h2>
                <button class="task-modal-close" onclick="this.closest('.task-modal-overlay').remove()">&times;</button>
            </div>
            <div class="task-modal-body">
                <div class="task-modal-main" style="width: 100%;">
                    <div class="task-section">
                        <label class="task-section-label">Название задачи *</label>
                        <input type="text" id="new-task-title" class="task-description-input" placeholder="Введите название задачи" required>
                    </div>
                    <div class="task-section">
                        <label class="task-section-label">Описание</label>
                        <textarea id="new-task-description" class="task-description-input" rows="6" placeholder="Введите описание задачи"></textarea>
                    </div>
                </div>
            </div>
            <div class="task-sidebar-actions" style="margin: 0; padding: var(--spacing-lg); border-top: 1px solid var(--border-light);">
                <button onclick="createTaskFromModal()" class="task-save-btn" style="width: 100%;">Создать задачу</button>
                <button onclick="this.closest('.task-modal-overlay').remove()" class="task-close-btn" style="width: 100%; margin-top: var(--spacing-sm);">Отмена</button>
            </div>
        </div>
    `;
    document.body.appendChild(modal);
    
    // Фокус на поле названия
    setTimeout(() => {
        const titleInput = document.getElementById('new-task-title');
        if (titleInput) titleInput.focus();
    }, 100);
}

function createTaskFromModal() {
    const API_BASE = window.AppConfig.API_BASE;
    const currentProjectId = window.AppConfig.getCurrentProjectId();
    const titleInput = document.getElementById('new-task-title');
    const descriptionInput = document.getElementById('new-task-description');
    
    if (!titleInput || !titleInput.value.trim()) {
        window.Utils.showMessage('Введите название задачи', 'error');
        return;
    }
    
    const title = titleInput.value.trim();
    const description = descriptionInput ? descriptionInput.value.trim() : '';
    
    // Закрываем модальное окно
    const modal = document.querySelector('.task-modal-overlay');
    if (modal) modal.remove();
    
    fetch(`${API_BASE}/tasks/projects/${currentProjectId}/tasks`, {
        method: 'POST',
        headers: { 
            'Content-Type': 'application/json', 
            'Authorization': `Bearer ${localStorage.getItem('token')}` 
        },
        body: JSON.stringify({ title, description })
    })
    .then(res => {
        if (!res.ok) {
            return res.json().then(data => {
                throw new Error(data.error || 'Ошибка создания задачи');
            });
        }
        return res.json();
    })
    .then(() => {
        window.Utils.showMessage('Задача создана', 'success');
        loadProjectTasks(currentProjectId);
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка создания задачи: ' + err.message, 'error');
    });
}

function viewTask(taskId) {
    const API_BASE = window.AppConfig.API_BASE;
    const currentProjectId = window.AppConfig.getCurrentProjectId();
    
    const currentUser = window.AppConfig.getCurrentUser();
    Promise.all([
        fetch(`${API_BASE}/tasks/tasks/${taskId}`, { headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` } })
            .then(res => { if (!res.ok) throw new Error('Task not found'); return res.json(); }),
        fetch(`${API_BASE}/tasks/tasks/${taskId}/comments`).then(res => res.ok ? res.json() : []),
        fetch(`${API_BASE}/projects/projects/${currentProjectId}`).then(res => res.ok ? res.json() : null),
        fetch(`${API_BASE}/projects/projects/${currentProjectId}/participants`).then(res => res.ok ? res.json() : [])
    ])
    .then(([task, comments, project, participants]) => {
        // Проверяем, что все данные корректны
        if (!task) {
            throw new Error('Задача не найдена');
        }
        
        const safeComments = Array.isArray(comments) ? comments : [];
        const safeParticipants = Array.isArray(participants) ? participants : [];
        
        const participantsMap = {};
        safeParticipants.forEach(p => {
            if (p && p.user_id) {
                participantsMap[p.user_id] = p;
            }
        });
        
        const assignedUser = task.assigned_to ? participantsMap[task.assigned_to] : null;
        const assigneeName = assignedUser ? (assignedUser.email ? assignedUser.email.split('@')[0] : `User ${task.assigned_to}`) : null;
        
        const isOrganizer = currentUser && project && project.organizer_id === currentUser.id;
        const userParticipation = currentUser ? safeParticipants.find(p => p.user_id === currentUser.id) : null;
        const isManager = userParticipation && (userParticipation.role === 'менеджер' || userParticipation.role === 'руководитель');
        const isLeader = userParticipation && userParticipation.role === 'руководитель';
        const canAssign = isOrganizer || isManager || isLeader;
        
        const statuses = ['новая', 'в работе', 'на проверке', 'завершена'];
        
        const modal = document.createElement('div');
        modal.className = 'task-modal-overlay';
        modal.onclick = (e) => { if (e.target === modal) modal.remove(); };
        modal.innerHTML = `
            <div class="task-modal-content">
                <div class="task-modal-header">
                    <input type="text" id="task-title-input" value="${window.Utils.escapeHtml(task.title)}" class="task-title-input">
                    <span class="task-modal-id">#${task.id}</span>
                    <button class="task-modal-close" onclick="this.closest('.task-modal-overlay').remove()">&times;</button>
                </div>
                <div class="task-modal-body">
                    <div class="task-modal-main">
                        <div class="task-section">
                            <label class="task-section-label">Описание</label>
                            <textarea id="task-description-input" class="task-description-input" rows="6">${window.Utils.escapeHtml(task.description || '')}</textarea>
                        </div>
                        <div class="task-section">
                            <label class="task-section-label">Комментарии (${safeComments.length})</label>
                            <div class="task-comments-list" id="task-comments-list">
                                ${safeComments.length > 0 ? safeComments.map(c => {
                                    if (!c || !c.user_id) return '';
                                    const commentUser = participantsMap[c.user_id];
                                    const commentUserName = commentUser ? (commentUser.email ? commentUser.email.split('@')[0] : `User ${c.user_id}`) : `User ${c.user_id}`;
                                    return `
                                    <div class="task-comment">
                                        <div class="task-comment-header">
                                            <div class="task-comment-author">${window.Utils.escapeHtml(commentUserName)}</div>
                                            <div class="task-comment-date">${c.publication_date ? new Date(c.publication_date).toLocaleString('ru-RU') : 'Неизвестно'}</div>
                                        </div>
                                        <div class="task-comment-content">${window.Utils.escapeHtml(c.content || '')}</div>
                                    </div>
                                `;
                                }).join('') : '<div class="task-comments-empty">Нет комментариев</div>'}
                            </div>
                            <form class="task-comment-form" onsubmit="addTaskComment(event, '${taskId}')">
                                <textarea id="task-comment-input" placeholder="Добавить комментарий..." required></textarea>
                                <button type="submit" class="task-comment-submit">Отправить</button>
                            </form>
                        </div>
                    </div>
                    <div class="task-modal-sidebar">
                        <div class="task-sidebar-section">
                            <label class="task-sidebar-label">Статус</label>
                            <select id="task-status-select" class="task-status-select">
                                ${statuses.map(s => `<option value="${s}" ${s === task.status ? 'selected' : ''}>${s}</option>`).join('')}
                            </select>
                        </div>
                        <div class="task-sidebar-section">
                            <label class="task-sidebar-label">Исполнитель</label>
                            ${canAssign ? `
                                <select id="task-assignee-select" class="task-assignee-select">
                                    <option value="">Не назначена</option>
                                    ${safeParticipants.map(p => {
                                        if (!p || !p.user_id) return '';
                                        const login = p.email ? p.email.split('@')[0] : `User ${p.user_id}`;
                                        return `<option value="${p.user_id}" ${p.user_id === task.assigned_to ? 'selected' : ''}>${window.Utils.escapeHtml(login)}</option>`;
                                    }).join('')}
                                </select>
                            ` : `
                                <div class="task-sidebar-value">${assigneeName || 'Не назначена'}</div>
                            `}
                        </div>
                        <div class="task-sidebar-section">
                            <label class="task-sidebar-label">Проект</label>
                            <div class="task-sidebar-value">${project ? window.Utils.escapeHtml(project.title) : 'Неизвестно'}</div>
                        </div>
                        <div class="task-sidebar-section">
                            <label class="task-sidebar-label">Дата создания</label>
                            <div class="task-sidebar-value">${new Date(task.creation_date).toLocaleDateString('ru-RU')}</div>
                        </div>
                        <div class="task-sidebar-actions">
                            <button onclick="saveTaskChanges('${taskId}')" class="task-save-btn">Сохранить изменения</button>
                            <button onclick="this.closest('.task-modal-overlay').remove()" class="task-close-btn">Закрыть</button>
                        </div>
                    </div>
                </div>
            </div>
        `;
        document.body.appendChild(modal);
        
        // Обработчики изменений
        document.getElementById('task-status-select').addEventListener('change', (e) => {
            updateTaskStatus(taskId, e.target.value);
        });
        const assigneeSelect = document.getElementById('task-assignee-select');
        if (assigneeSelect) {
            assigneeSelect.addEventListener('change', (e) => {
                assignTask(taskId, e.target.value || null);
            });
        }
    })
    .catch(err => { 
        window.Utils.showMessage('Ошибка загрузки задачи: ' + err.message, 'error'); 
    });
}

function saveTaskChanges(taskId) {
    const API_BASE = window.AppConfig.API_BASE;
    const title = document.getElementById('task-title-input').value;
    const description = document.getElementById('task-description-input').value;
    
    fetch(`${API_BASE}/tasks/tasks/${taskId}`, {
        method: 'PUT',
        headers: { 
            'Content-Type': 'application/json', 
            'Authorization': `Bearer ${localStorage.getItem('token')}` 
        },
        body: JSON.stringify({ title, description })
    })
    .then(res => {
        if (!res.ok) throw new Error('Failed to update task');
        window.Utils.showMessage('Задача обновлена', 'success');
        const currentProjectId = window.AppConfig.getCurrentProjectId();
        loadProjectTasks(currentProjectId);
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка обновления задачи: ' + err.message, 'error');
    });
}

function assignTask(taskId, userId) {
    const API_BASE = window.AppConfig.API_BASE;
    fetch(`${API_BASE}/tasks/tasks/${taskId}/assign`, {
        method: 'POST',
        headers: { 
            'Content-Type': 'application/json', 
            'Authorization': `Bearer ${localStorage.getItem('token')}` 
        },
        body: JSON.stringify({ assigned_to: userId || null })
    })
    .then(res => {
        if (!res.ok) throw new Error('Failed to assign task');
        window.Utils.showMessage('Исполнитель обновлен', 'success');
        const currentProjectId = window.AppConfig.getCurrentProjectId();
        loadProjectTasks(currentProjectId);
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка назначения исполнителя: ' + err.message, 'error');
    });
}

function addTaskComment(event, taskId) {
    event.preventDefault();
    const API_BASE = window.AppConfig.API_BASE;
    const content = document.getElementById('task-comment-input').value;
    if (!content) return;
    fetch(`${API_BASE}/tasks/tasks/${taskId}/comments`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', 'Authorization': `Bearer ${localStorage.getItem('token')}` },
        body: JSON.stringify({ content })
    })
    .then(res => { if (!res.ok) throw new Error('Failed to add comment'); window.Utils.showMessage('Комментарий добавлен', 'success'); viewTask(taskId); })
    .catch(err => { window.Utils.showMessage('Ошибка добавления комментария: ' + err.message, 'error'); });
}

function updateTaskStatusPrompt(taskId) {
    const API_BASE = window.AppConfig.API_BASE;
    const currentProjectId = window.AppConfig.getCurrentProjectId();
    const statuses = ['новая', 'в работе', 'на проверке', 'завершена'];
    const currentStatus = prompt('Выберите статус:\n' + statuses.map((s, i) => `${i + 1}. ${s}`).join('\n'));
    if (!currentStatus) return;
    const statusIndex = parseInt(currentStatus) - 1;
    if (statusIndex < 0 || statusIndex >= statuses.length) { window.Utils.showMessage('Неверный статус', 'error'); return; }
    updateTaskStatus(taskId, statuses[statusIndex]);
}

// Drag and Drop функции для канбан-доски
let draggedTaskId = null;
let draggedTaskStatus = null;

function handleDragStart(event) {
    draggedTaskId = event.target.dataset.taskId;
    draggedTaskStatus = event.target.dataset.status;
    event.target.style.opacity = '0.5';
}

function handleDragOver(event) {
    event.preventDefault();
    const column = event.currentTarget;
    column.style.backgroundColor = 'var(--primary-bg)';
}

function handleDrop(event) {
    event.preventDefault();
    const column = event.currentTarget;
    const newStatus = column.dataset.status;
    column.style.backgroundColor = '';
    
    if (draggedTaskId && newStatus && newStatus !== draggedTaskStatus) {
        updateTaskStatus(draggedTaskId, newStatus);
    }
    
    // Сбрасываем стили
    document.querySelectorAll('.kanban-task-card').forEach(card => {
        card.style.opacity = '1';
    });
    document.querySelectorAll('.kanban-column').forEach(col => {
        col.style.backgroundColor = '';
    });
}

function updateTaskStatus(taskId, newStatus) {
    const API_BASE = window.AppConfig.API_BASE;
    const currentProjectId = window.AppConfig.getCurrentProjectId();
    
    fetch(`${API_BASE}/tasks/tasks/${taskId}/status`, {
        method: 'PUT',
        headers: { 
            'Content-Type': 'application/json', 
            'Authorization': `Bearer ${localStorage.getItem('token')}` 
        },
        body: JSON.stringify({ status: newStatus })
    })
    .then(res => {
        if (!res.ok) throw new Error('Failed to update status');
        return res.json();
    })
    .then(() => {
        window.Utils.showMessage('Статус задачи обновлен', 'success');
        loadProjectTasks(currentProjectId);
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка обновления статуса: ' + err.message, 'error');
        loadProjectTasks(currentProjectId);
    });
}

window.showProjectTasks = function(projectId) {
    window.AppConfig.setCurrentProjectId(projectId);
    window.viewProject(projectId);
    setTimeout(() => {
        window.showProjectDetailTab('tasks');
        loadProjectTasks(projectId);
    }, 300);
};

// Экспорт функций
window.Tasks = {
    loadProjectTasks,
    showCreateTaskForm,
    createTaskFromModal,
    viewTask,
    addTaskComment,
    updateTaskStatus,
    updateTaskStatusPrompt,
    saveTaskChanges,
    assignTask
};

// Глобальные функции для использования в HTML
window.viewTask = viewTask;
window.addTaskComment = addTaskComment;
window.updateTaskStatusPrompt = updateTaskStatusPrompt;
window.saveTaskChanges = saveTaskChanges;
window.assignTask = assignTask;
window.createTaskFromModal = createTaskFromModal;
window.handleDragStart = handleDragStart;
window.handleDragOver = handleDragOver;
window.handleDrop = handleDrop;

