# Структура JavaScript модулей

Код разделен на логические модули для удобства редактирования и поддержки.

## Структура модулей

### 1. `config.js` - Конфигурация
- Глобальные переменные (API_BASE, currentUser, currentRole, currentProjectId)
- Настройки доступности
- Экспорт через `window.AppConfig`

### 2. `utils.js` - Утилиты
- `showMessage()` - показ сообщений
- `escapeHtml()` - экранирование HTML
- `showPage()` - переключение страниц
- `updateUI()` - обновление интерфейса
- Экспорт через `window.Utils`

### 3. `auth.js` - Авторизация
- `checkAuth()` - проверка авторизации
- `login()` - вход
- `register()` - регистрация
- `logout()` - выход
- `loadProfile()` - загрузка профиля
- `submitOrganizerRequest()` - заявка на роль организатора
- Функции редактирования профиля
- Экспорт через `window.Auth`

### 4. `projects.js` - Проекты
- `loadProjects()` - загрузка списка проектов
- `createProject()` - создание проекта
- `loadMyProjects()` - загрузка моих проектов
- `manageProject()` - управление проектом
- `submitParticipationRequest()` - заявка на участие
- `loadCategories()` - загрузка категорий
- Управление тегами проектов
- `viewProject()` - просмотр проекта
- Экспорт через `window.Projects`

### 5. `tasks.js` - Задачи
- `loadProjectTasks()` - загрузка задач проекта
- `showCreateTaskForm()` - форма создания задачи
- `viewTask()` - просмотр задачи
- `addTaskComment()` - добавление комментария
- `updateTaskStatus()` - обновление статуса
- `showProjectTasks()` - показ задач проекта
- Экспорт через `window.Tasks`

### 6. `chat.js` - Чат
- `loadProjectChat()` - загрузка чата проекта
- `createProjectChat()` - создание чата
- `loadChatMessages()` - загрузка сообщений
- `sendChatMessage()` - отправка сообщения
- `connectWebSocket()` - подключение WebSocket
- `loadChatsPage()` - страница чатов
- `loadProjectChats()` - загрузка чатов проекта
- Экспорт через `window.Chat`

### 7. `admin.js` - Админ панель
- `showAdminTab()` - переключение вкладок
- `loadUsers()` - загрузка пользователей
- `toggleUserStatus()` - изменение статуса пользователя
- `loadOrganizerRequests()` - загрузка заявок организаторов
- `approveOrganizerRequest()` - одобрение заявки
- `rejectOrganizerRequest()` - отклонение заявки
- `loadReports()` - загрузка отчетов
- Экспорт через `window.Admin`

### 8. `accessibility.js` - Доступность
- `toggleAccessibility()` - показ/скрытие панели
- `changeFontSize()` - изменение размера шрифта
- `toggleHighContrast()` - высокий контраст
- `toggleSimpleNav()` - упрощенная навигация
- `toggleLargeButtons()` - увеличенные кнопки
- `resetAccessibility()` - сброс настроек
- `applyAccessibilitySettings()` - применение настроек
- `saveAccessibilitySettings()` - сохранение настроек
- `loadAccessibilitySettings()` - загрузка настроек
- Экспорт через `window.Accessibility`

### 9. `app.js` - Главный файл
- `init()` - инициализация приложения
- Загружает все модули и запускает приложение

## Порядок загрузки модулей

Модули должны загружаться в следующем порядке:
1. `config.js` - конфигурация
2. `utils.js` - утилиты
3. `accessibility.js` - доступность
4. `auth.js` - авторизация
5. `projects.js` - проекты
6. `tasks.js` - задачи
7. `chat.js` - чат
8. `admin.js` - админ панель
9. `app.js` - инициализация

## Использование

Все функции доступны через соответствующие объекты:
- `window.AppConfig` - конфигурация
- `window.Utils` - утилиты
- `window.Auth` - авторизация
- `window.Projects` - проекты
- `window.Tasks` - задачи
- `window.Chat` - чат
- `window.Admin` - админ панель
- `window.Accessibility` - доступность

Некоторые функции также доступны глобально для использования в HTML (onclick, onsubmit).

