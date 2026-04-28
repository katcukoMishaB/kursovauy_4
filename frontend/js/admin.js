// Модуль админ панели

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
    const API_BASE = window.AppConfig.API_BASE;
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
            window.Utils.showMessage('Ошибка: неверный формат данных пользователей', 'error');
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
                            <td>${window.Utils.escapeHtml(user.first_name || '')}</td>
                            <td>${window.Utils.escapeHtml(user.last_name || '')}</td>
                            <td>${window.Utils.escapeHtml(user.email || '')}</td>
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
        window.Utils.showMessage('Ошибка загрузки пользователей: ' + err.message, 'error');
    });
}

function toggleUserStatus(userId, status) {
    const API_BASE = window.AppConfig.API_BASE;
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
        window.Utils.showMessage('Статус пользователя обновлен', 'success');
        loadUsers();
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка обновления статуса: ' + err.message, 'error');
    });
}

function loadOrganizerRequests() {
    const API_BASE = window.AppConfig.API_BASE;
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
            window.Utils.showMessage('Ошибка: неверный формат данных заявок', 'error');
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
                            <td>${window.Utils.escapeHtml(req.experience_description || '')}</td>
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
        window.Utils.showMessage('Ошибка загрузки заявок: ' + err.message, 'error');
    });
}

function approveOrganizerRequest(requestId) {
    const API_BASE = window.AppConfig.API_BASE;
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
        window.Utils.showMessage('Заявка одобрена', 'success');
        loadOrganizerRequests();
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка одобрения заявки: ' + err.message, 'error');
    });
}

function rejectOrganizerRequest(requestId) {
    const API_BASE = window.AppConfig.API_BASE;
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
        window.Utils.showMessage('Заявка отклонена', 'success');
        loadOrganizerRequests();
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка отклонения заявки: ' + err.message, 'error');
    });
}

function loadReports() {
    const API_BASE = window.AppConfig.API_BASE;
    const content = document.getElementById('admin-content');
    if (!content) return;
    
    Promise.all([
        fetch(`${API_BASE}/projects/projects`, {
            headers: {
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            }
        }).then(res => res.ok ? res.json() : []).catch(() => [])
    ])
    .then(([projects]) => {
        content.innerHTML = `
            <div class="form-container">
                <h3>Генерация отчетов</h3>
                <div style="margin-bottom: var(--spacing-lg);">
                    <h4 style="margin-bottom: var(--spacing);">Отчет по проекту</h4>
                    <select id="admin-project-select" style="width: 100%; padding: var(--spacing-sm); margin-bottom: var(--spacing); border-radius: var(--radius); border: 1px solid var(--border);">
                        <option value="">Выберите проект</option>
                        ${projects.map(p => `<option value="${p.id}">${window.Utils.escapeHtml(p.title || 'Без названия')}</option>`).join('')}
                    </select>
                    <button onclick="generateAdminProjectReport()" style="width: 100%; padding: var(--spacing-sm); margin-bottom: var(--spacing-lg);">Скачать отчет по проекту</button>
                </div>
                <div>
                    <h4 style="margin-bottom: var(--spacing);">Отчет по всем проектам</h4>
                    <button onclick="generateAllProjectsReport()" style="width: 100%; padding: var(--spacing-sm);">Скачать отчет по всем проектам</button>
                </div>
            </div>
        `;
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка загрузки проектов: ' + err.message, 'error');
    });
}

function generateAdminProjectReport() {
    const API_BASE = window.AppConfig.API_BASE;
    const projectId = document.getElementById('admin-project-select')?.value;
    if (!projectId) {
        window.Utils.showMessage('Выберите проект', 'error');
        return;
    }
    const token = localStorage.getItem('token');
    
    fetch(`${API_BASE}/reports/excel/admin/project/${projectId}`, {
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
        const link = document.createElement('a');
        link.href = url;
        link.download = `admin_project_report_${projectId}.xlsx`;
        link.click();
        window.URL.revokeObjectURL(url);
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка генерации отчета: ' + err.message, 'error');
    });
}

function generateAllProjectsReport() {
    const API_BASE = window.AppConfig.API_BASE;
    const token = localStorage.getItem('token');
    
    fetch(`${API_BASE}/reports/excel/all-projects`, {
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
        const link = document.createElement('a');
        link.href = url;
        link.download = `all_projects_report_${new Date().toISOString().split('T')[0]}.xlsx`;
        link.click();
        window.URL.revokeObjectURL(url);
    })
    .catch(err => {
        window.Utils.showMessage('Ошибка генерации отчета: ' + err.message, 'error');
    });
}

// Экспорт функций
window.Admin = {
    showAdminTab,
    loadUsers,
    toggleUserStatus,
    loadOrganizerRequests,
    approveOrganizerRequest,
    rejectOrganizerRequest,
    loadReports,
    generateAdminProjectReport,
    generateAllProjectsReport
};

// Глобальные функции для использования в HTML
window.showAdminTab = showAdminTab;
window.toggleUserStatus = toggleUserStatus;
window.approveOrganizerRequest = approveOrganizerRequest;
window.rejectOrganizerRequest = rejectOrganizerRequest;

