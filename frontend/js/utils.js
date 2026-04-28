// Утилиты и вспомогательные функции

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

function showPage(pageId) {
    const currentUser = window.AppConfig.getCurrentUser();
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
    
    // Загрузка данных для конкретных страниц
    if (pageId === 'projects') {
        if (window.Projects && window.Projects.loadProjects) window.Projects.loadProjects();
    } else if (pageId === 'profile') {
        if (window.Auth && window.Auth.loadProfile) window.Auth.loadProfile();
    } else if (pageId === 'admin') {
        if (window.Admin && window.Admin.showAdminTab) window.Admin.showAdminTab('users');
    } else if (pageId === 'my-projects') {
        if (window.Projects && window.Projects.loadMyProjectsPage) window.Projects.loadMyProjectsPage();
    } else if (pageId === 'chats') {
        if (window.Chat && window.Chat.loadChatsPage) window.Chat.loadChatsPage();
    }
}

function updateUI() {
    const currentUser = window.AppConfig.getCurrentUser();
    const currentRole = window.AppConfig.getCurrentRole();
    const header = document.getElementById('main-header');
    if (currentUser) {
        if (header) header.style.display = 'block';
        const logoutBtn = document.getElementById('logout-btn');
        if (logoutBtn) logoutBtn.style.display = 'block';
        const profileLink = document.getElementById('profile-link');
        if (profileLink) profileLink.style.display = 'block';
        const myProjectsLink = document.getElementById('my-projects-link');
        if (myProjectsLink) myProjectsLink.style.display = 'block';
        const chatsLink = document.getElementById('chats-link');
        if (chatsLink) chatsLink.style.display = 'block';
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

function showProjectDetailTab(tabName) {
    document.querySelectorAll('.project-tab').forEach(tab => tab.classList.remove('active'));
    document.querySelectorAll('.project-tab-content').forEach(content => {
        content.classList.remove('active');
        content.style.display = 'none';
    });
    
    const activeTab = document.querySelector(`.project-tab[onclick*="${tabName}"]`);
    const activeContent = document.getElementById(`project-${tabName}-tab`);
    
    if (activeTab && activeTab.style.display !== 'none') {
        activeTab.classList.add('active');
        activeTab.style.display = 'inline-block';
    }
    if (activeContent) {
        activeContent.classList.add('active');
        activeContent.style.display = 'block';
    }
    
    // Загружаем данные для вкладки, если нужно
    if (tabName === 'tasks') {
        const currentProjectId = window.AppConfig.getCurrentProjectId();
        if (currentProjectId) {
            window.Tasks.loadProjectTasks(currentProjectId);
        }
    } else if (tabName === 'chat') {
        const currentProjectId = window.AppConfig.getCurrentProjectId();
        if (currentProjectId) {
            window.Chat.loadProjectChatMessages(currentProjectId);
        }
    }
}

// Экспорт функций
window.Utils = {
    showMessage,
    escapeHtml,
    showPage,
    updateUI
};

// Глобальные функции для обратной совместимости
window.showMessage = showMessage;
window.escapeHtml = escapeHtml;
window.showPage = showPage;
window.showProjectDetailTab = showProjectDetailTab;
