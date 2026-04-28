// Модуль настроек доступности

function toggleAccessibility() {
    const panel = document.getElementById('accessibility-panel');
    panel.style.display = panel.style.display === 'none' ? 'block' : 'none';
}

function changeFontSize(size) {
    const accessibilitySettings = window.AppConfig.accessibilitySettings;
    accessibilitySettings.fontSize = size;
    
    // Удаляем все классы размера шрифта
    document.body.classList.remove('font-size-default', 'font-size-large', 'font-size-xlarge');
    
    // Добавляем новый класс
    if (size !== 'default') {
        document.body.classList.add(`font-size-${size}`);
    }
    
    saveAccessibilitySettings();
}

function toggleHighContrast(enabled) {
    const accessibilitySettings = window.AppConfig.accessibilitySettings;
    accessibilitySettings.highContrast = enabled;
    if (enabled) {
        document.body.classList.add('high-contrast');
    } else {
        document.body.classList.remove('high-contrast');
    }
    saveAccessibilitySettings();
}


function changeColorScheme(scheme) {
    const accessibilitySettings = window.AppConfig.accessibilitySettings;
    accessibilitySettings.colorScheme = scheme;
    document.body.className = document.body.className.replace(/color-scheme-\w+/g, '');
    if (scheme !== 'default') {
        document.body.classList.add(`color-scheme-${scheme}`);
    }
    saveAccessibilitySettings();
}

function toggleImages(enabled) {
    const accessibilitySettings = window.AppConfig.accessibilitySettings;
    accessibilitySettings.imagesEnabled = enabled;
    if (enabled) {
        document.body.classList.remove('images-disabled');
    } else {
        document.body.classList.add('images-disabled');
    }
    saveAccessibilitySettings();
}

function toggleImagesGrayscale(enabled) {
    const accessibilitySettings = window.AppConfig.accessibilitySettings;
    accessibilitySettings.imagesGrayscale = enabled;
    if (enabled) {
        document.body.classList.add('images-grayscale');
    } else {
        document.body.classList.remove('images-grayscale');
    }
    saveAccessibilitySettings();
}

function resetAccessibility() {
    const accessibilitySettings = window.AppConfig.accessibilitySettings;
    accessibilitySettings.fontSize = 'default';
    accessibilitySettings.highContrast = false;
    accessibilitySettings.colorScheme = 'default';
    accessibilitySettings.imagesEnabled = true;
    accessibilitySettings.imagesGrayscale = false;
    applyAccessibilitySettings();
    saveAccessibilitySettings();
    const fontSizeSelect = document.getElementById('font-size-select');
    const highContrast = document.getElementById('high-contrast');
    const colorScheme = document.getElementById('color-scheme-select');
    const imagesEnabled = document.getElementById('images-enabled');
    const imagesGrayscale = document.getElementById('images-grayscale');
    if (fontSizeSelect) fontSizeSelect.value = 'default';
    if (highContrast) highContrast.checked = false;
    if (colorScheme) colorScheme.value = 'default';
    if (imagesEnabled) imagesEnabled.checked = true;
    if (imagesGrayscale) imagesGrayscale.checked = false;
}

function applyAccessibilitySettings() {
    const accessibilitySettings = window.AppConfig.accessibilitySettings;
    
    const fontSizeSelect = document.getElementById('font-size-select');
    const highContrast = document.getElementById('high-contrast');
    const colorScheme = document.getElementById('color-scheme-select');
    const imagesEnabled = document.getElementById('images-enabled');
    const imagesGrayscale = document.getElementById('images-grayscale');
    
    if (fontSizeSelect) fontSizeSelect.value = accessibilitySettings.fontSize;
    if (highContrast) highContrast.checked = accessibilitySettings.highContrast;
    if (colorScheme) colorScheme.value = accessibilitySettings.colorScheme;
    if (imagesEnabled) imagesEnabled.checked = accessibilitySettings.imagesEnabled;
    if (imagesGrayscale) imagesGrayscale.checked = accessibilitySettings.imagesGrayscale;
    
    changeFontSize(accessibilitySettings.fontSize);
    if (accessibilitySettings.highContrast) {
        document.body.classList.add('high-contrast');
    } else {
        document.body.classList.remove('high-contrast');
    }
    changeColorScheme(accessibilitySettings.colorScheme);
    toggleImages(accessibilitySettings.imagesEnabled);
    toggleImagesGrayscale(accessibilitySettings.imagesGrayscale);
}

function saveAccessibilitySettings() {
    const accessibilitySettings = window.AppConfig.accessibilitySettings;
    localStorage.setItem('accessibilitySettings', JSON.stringify(accessibilitySettings));
}

function loadAccessibilitySettings() {
    const accessibilitySettings = window.AppConfig.accessibilitySettings;
    const saved = localStorage.getItem('accessibilitySettings');
    if (saved) {
        Object.assign(accessibilitySettings, JSON.parse(saved));
    }
}

// Экспорт функций
window.Accessibility = {
    toggleAccessibility,
    changeFontSize,
    toggleHighContrast,
    changeColorScheme,
    toggleImages,
    toggleImagesGrayscale,
    resetAccessibility,
    applyAccessibilitySettings,
    saveAccessibilitySettings,
    loadAccessibilitySettings
};

