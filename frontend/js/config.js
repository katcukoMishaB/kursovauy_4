// Конфигурация и глобальные переменные
const API_BASE = '/api';

// Глобальное состояние приложения
let currentUser = null;
let currentRole = null;
let currentProjectId = null;

// Настройки доступности
const accessibilitySettings = {
    fontSize: 'default', // 'default', 'large', 'xlarge'
    highContrast: false,
    colorScheme: 'default', // 'default', 'protanopia', 'deuteranopia', 'tritanopia', 'grayscale'
    imagesEnabled: true,
    imagesGrayscale: false
};

// Экспорт для использования в других модулях
window.AppConfig = {
    API_BASE,
    getCurrentUser: () => currentUser,
    setCurrentUser: (user) => { currentUser = user; },
    getCurrentRole: () => currentRole,
    setCurrentRole: (role) => { currentRole = role; },
    getCurrentProjectId: () => currentProjectId,
    setCurrentProjectId: (id) => { currentProjectId = id; },
    accessibilitySettings
};

