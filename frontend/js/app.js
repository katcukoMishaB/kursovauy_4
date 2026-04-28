// Главный файл приложения - инициализация всех модулей

function init() {
    // Загружаем настройки доступности
    window.Accessibility.loadAccessibilitySettings();
    window.Accessibility.applyAccessibilitySettings();
    
    // Проверяем авторизацию
    window.Auth.checkAuth();
    
    // Загружаем категории проектов
    window.Projects.loadCategories();
}

// Инициализация при загрузке страницы
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', init);
} else {
    init();
}

