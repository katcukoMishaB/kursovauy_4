// Модуль чата

let currentChatWS = null;
let currentChatId = null;
let isLoadingChats = false;

function loadProjectChat(projectId) {
    const API_BASE = window.AppConfig.API_BASE;
    window.AppConfig.setCurrentProjectId(projectId);
    window.viewProject(projectId);
    setTimeout(() => {
        window.showProjectDetailTab('chat');
        loadProjectChatMessages(projectId);
    }, 300);
}

function loadProjectChatMessages(projectId) {
    const API_BASE = window.AppConfig.API_BASE;
    const chatMessages = document.querySelector('#project-chat-tab #chat-messages');
    if (!chatMessages) {
        setTimeout(() => loadProjectChatMessages(projectId), 100);
        return;
    }
    
    chatMessages.innerHTML = '<p style="text-align: center; padding: var(--spacing-xl); color: var(--text-secondary);">Загрузка чата...</p>';
    
    fetch(`${API_BASE}/chats/projects/${projectId}/chats`, { 
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` } 
    })
    .then(res => res.ok ? res.json() : [])
    .then(chats => {
        if (!chats || !Array.isArray(chats) || chats.length === 0) { 
            createProjectChat(projectId); 
            return; 
        }
        currentChatId = chats[0].id;
        loadChatMessages(chats[0].id);
        connectWebSocket(chats[0].id);
    })
    .catch(err => { 
        console.error('Error loading chat:', err); 
        if (chatMessages) {
            chatMessages.innerHTML = '<p style="text-align: center; padding: var(--spacing-xl); color: var(--danger);">Ошибка загрузки чата</p>';
        }
    });
}

function createProjectChat(projectId) {
    const API_BASE = window.AppConfig.API_BASE;
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
        window.AppConfig.setCurrentProjectId(projectId);
        
        // Обновляем интерфейс чата
        const chatMessages = document.querySelector('#project-chat-tab #chat-messages');
        const content = document.getElementById('chats-content');
        
        if (chatMessages) {
            // Если мы на странице проекта
            setTimeout(() => {
                loadChatMessages(data.id);
                connectWebSocket(data.id);
            }, 100);
        } else if (content) {
            // Если мы на странице чатов - создаем такую же структуру как в проекте
            content.innerHTML = `
                <div id="standalone-chat-tab" class="project-tab-content" style="display: block;">
                    <div class="project-chat-container">
                        <div id="standalone-chat-messages" class="chat-messages-container scrollable-content"></div>
                        <div id="standalone-chat-input-container" class="chat-input-container">
                            <input type="text" id="standalone-chat-message-input" placeholder="Введите сообщение..." onkeypress="if(event.key==='Enter') { event.preventDefault(); window.Chat.sendStandaloneChatMessage(); }">
                            <button onclick="window.Chat.sendStandaloneChatMessage()" class="send-button">
                                <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
                                    <path d="M18 2L9 11M18 2L12 18L9 11M18 2L2 8L9 11" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                                </svg>
                            </button>
                        </div>
                    </div>
                </div>
            `;
            // Загружаем сообщения после создания элементов
            setTimeout(() => {
                const messagesEl = document.getElementById('standalone-chat-messages');
                if (messagesEl) {
                    loadStandaloneChatMessages(data.id);
                    connectWebSocket(data.id);
                } else {
                    console.error('Chat messages container not found in createProjectChat');
                }
            }, 100);
        }
    })
    .catch(err => { 
        console.error('Error creating chat:', err);
        window.Utils.showMessage('Ошибка создания чата: ' + err.message, 'error');
        const content = document.getElementById('chats-content');
        if (content) content.innerHTML = '<p>Ошибка создания чата</p>';
    });
}

function loadChatMessages(chatId) {
    const API_BASE = window.AppConfig.API_BASE;
    const currentUser = window.AppConfig.getCurrentUser();
    if (!chatId) {
        console.error('[CHAT] loadChatMessages: chatId is null');
        return;
    }
    
    console.log('[CHAT] Loading messages for chat:', chatId);
    
    fetch(`${API_BASE}/chats/chats/${chatId}/messages`, { 
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` } 
    })
    .then(res => {
        console.log('[CHAT] Fetch response status:', res.status);
        if (!res.ok) {
            console.error('[CHAT] Failed to load messages:', res.status);
            return [];
        }
        return res.json();
    })
    .then(messages => {
        console.log('[CHAT] Messages received:', messages);
        console.log('[CHAT] Messages count:', messages ? messages.length : 0);
        
        // Ищем контейнер сообщений - сначала по ID, потом по селекторам
        let chatMessages = document.getElementById('chat-messages');
        console.log('[CHAT] Container by ID:', !!chatMessages);
        
        if (!chatMessages) {
            chatMessages = document.querySelector('#project-chat-tab #chat-messages');
            console.log('[CHAT] Container by project tab:', !!chatMessages);
        }
        if (!chatMessages) {
            chatMessages = document.querySelector('.chat-messages-container');
            console.log('[CHAT] Container by class:', !!chatMessages);
        }
        if (!chatMessages) {
            console.error('[CHAT] Container not found! Searching all elements...');
            console.log('[CHAT] All elements with chat:', document.querySelectorAll('[id*="chat"], [class*="chat"]'));
            console.log('[CHAT] chats-content element:', document.getElementById('chats-content'));
            console.log('[CHAT] project-chat-container:', document.querySelector('.project-chat-container'));
            
            // Попробуем еще раз через небольшую задержку
            setTimeout(() => {
                let retryChatMessages = document.getElementById('chat-messages');
                if (!retryChatMessages) {
                    retryChatMessages = document.querySelector('.chat-messages-container');
                }
                console.log('[CHAT] Retry - container found:', !!retryChatMessages);
                if (retryChatMessages && messages && Array.isArray(messages)) {
                    if (messages.length === 0) {
                        retryChatMessages.innerHTML = '<p style="text-align: center; color: var(--text-secondary); padding: var(--spacing-xl);">Нет сообщений</p>';
                    } else {
                        console.log('[CHAT] Rendering', messages.length, 'messages on retry');
                        retryChatMessages.innerHTML = messages.map(msg => `
                            <div class="chat-message ${msg.user_id === currentUser.id ? 'own' : ''}">
                                <div class="message-header">
                                    <strong>${msg.user_id === currentUser.id ? 'Вы' : 'Участник'}</strong>
                                    <span class="message-time">${new Date(msg.sent_at).toLocaleString('ru-RU')}</span>
                                </div>
                                <div class="message-text">${window.Utils.escapeHtml(msg.message_text)}</div>
                            </div>
                        `).join('');
                        retryChatMessages.scrollTop = retryChatMessages.scrollHeight;
                        console.log('[CHAT] Messages rendered on retry');
                    }
                }
            }, 200);
            return;
        }
        
        if (!messages || !Array.isArray(messages)) {
            console.warn('[CHAT] Messages is not an array:', messages);
            messages = [];
        }
        
        if (messages.length === 0) {
            console.log('[CHAT] No messages, showing empty state');
            chatMessages.innerHTML = '<p style="text-align: center; color: var(--text-secondary); padding: var(--spacing-xl);">Нет сообщений</p>';
            return;
        }
        
        console.log('[CHAT] Rendering', messages.length, 'messages into container');
        console.log('[CHAT] Container before render:', {
            innerHTML: chatMessages.innerHTML.substring(0, 100),
            scrollHeight: chatMessages.scrollHeight,
            clientHeight: chatMessages.clientHeight,
            offsetHeight: chatMessages.offsetHeight,
            style: window.getComputedStyle(chatMessages).display
        });
        
        const messagesHTML = messages.map(msg => `
            <div class="chat-message ${msg.user_id === currentUser.id ? 'own' : ''}">
                <div class="message-header">
                    <strong>${msg.user_id === currentUser.id ? 'Вы' : 'Участник'}</strong>
                    <span class="message-time">${new Date(msg.sent_at).toLocaleString('ru-RU')}</span>
                </div>
                <div class="message-text">${window.Utils.escapeHtml(msg.message_text)}</div>
            </div>
        `).join('');
        
        chatMessages.innerHTML = messagesHTML;
        
        // Принудительно устанавливаем высоту контейнера после рендеринга
        requestAnimationFrame(() => {
            setTimeout(() => {
                const content = document.getElementById('chats-content');
                const container = content ? content.querySelector('.project-chat-container') : null;
                
                if (container && chatMessages) {
                    const containerHeight = container.clientHeight || container.offsetHeight;
                    const inputHeight = 60; // примерная высота input контейнера
                    
                    if (containerHeight > 0) {
                        chatMessages.style.height = (containerHeight - inputHeight) + 'px';
                        chatMessages.style.minHeight = (containerHeight - inputHeight) + 'px';
                        console.log('[CHAT] Set explicit height for messages container:', containerHeight - inputHeight);
                    } else {
                        // Если контейнер все еще не имеет высоты, используем высоту родителя
                        const parentHeight = content ? content.clientHeight : 0;
                        if (parentHeight > 0) {
                            container.style.height = parentHeight + 'px';
                            chatMessages.style.height = (parentHeight - inputHeight) + 'px';
                            chatMessages.style.minHeight = (parentHeight - inputHeight) + 'px';
                            console.log('[CHAT] Set explicit height from parent:', parentHeight);
                        }
                    }
                }
                
                chatMessages.scrollTop = chatMessages.scrollHeight;
                console.log('[CHAT] Messages rendered successfully');
                console.log('[CHAT] Container after render:', {
                    scrollHeight: chatMessages.scrollHeight,
                    clientHeight: chatMessages.clientHeight,
                    offsetHeight: chatMessages.offsetHeight,
                    scrollTop: chatMessages.scrollTop,
                    childrenCount: chatMessages.children.length,
                    styleHeight: chatMessages.style.height,
                    computedHeight: window.getComputedStyle(chatMessages).height
                });
            }, 150);
        });
    })
    .catch(err => { 
        console.error('[CHAT] Error loading messages:', err);
        const chatMessages = document.querySelector('#project-chat-tab #chat-messages') || 
                            document.querySelector('.chat-messages-container') ||
                            document.getElementById('chat-messages');
        if (chatMessages) {
            chatMessages.innerHTML = '<p style="color: var(--danger); text-align: center; padding: var(--spacing-xl);">Ошибка загрузки сообщений</p>';
        }
    });
}

function connectWebSocket(chatId) {
    if (currentChatWS) currentChatWS.close();
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/api/chats/chats/${chatId}/ws`;
    console.log('[CHAT] Connecting WebSocket to:', wsUrl);
    currentChatWS = new WebSocket(wsUrl);
    currentChatWS.onopen = () => { 
        console.log('[CHAT] WebSocket connected'); 
    };
    currentChatWS.onmessage = (event) => { 
        console.log('[CHAT] WebSocket message received:', event.data);
        try {
            const message = JSON.parse(event.data);
            console.log('[CHAT] Parsed message:', message);
            addMessageToChat(message);
        } catch (err) {
            console.error('[CHAT] Error parsing WebSocket message:', err, event.data);
        }
    };
    currentChatWS.onerror = (error) => { 
        console.error('[CHAT] WebSocket error:', error); 
    };
    currentChatWS.onclose = () => { 
        console.log('[CHAT] WebSocket disconnected'); 
        setTimeout(() => { 
            if (currentChatId) {
                console.log('[CHAT] Reconnecting WebSocket...');
                connectWebSocket(currentChatId); 
            }
        }, 3000); 
    };
}

function addMessageToChat(message) {
    console.log('[CHAT] addMessageToChat called with message:', message);
    const currentUser = window.AppConfig.getCurrentUser();
    if (!currentUser) {
        console.error('[CHAT] No current user found');
        return;
    }
    
    // Сначала пробуем найти контейнер в standalone чате
    let chatMessages = document.getElementById('standalone-chat-messages');
    console.log('[CHAT] standalone-chat-messages found:', !!chatMessages);
    
    // Потом в проекте
    if (!chatMessages) {
        chatMessages = document.querySelector('#project-chat-tab #chat-messages');
        console.log('[CHAT] project-chat-tab #chat-messages found:', !!chatMessages);
    }
    
    // И наконец общий поиск
    if (!chatMessages) {
        chatMessages = document.getElementById('chat-messages');
        console.log('[CHAT] chat-messages by ID found:', !!chatMessages);
    }
    if (!chatMessages) {
        chatMessages = document.querySelector('.chat-messages-container');
        console.log('[CHAT] .chat-messages-container found:', !!chatMessages);
    }
    if (!chatMessages) {
        console.error('[CHAT] Chat messages container not found in addMessageToChat');
        return;
    }
    
    // Проверяем формат сообщения
    if (!message || !message.message_text) {
        console.error('[CHAT] Invalid message format:', message);
        return;
    }
    
    const messageDiv = document.createElement('div');
    messageDiv.className = `chat-message ${message.user_id === currentUser.id ? 'own' : ''}`;
    messageDiv.innerHTML = `<div class="message-header"><strong>${message.user_id === currentUser.id ? 'Вы' : 'Участник'}</strong>
    <span class="message-time">${new Date(message.sent_at || Date.now()).toLocaleString('ru-RU')}</span></div>
    <div class="message-text">${window.Utils.escapeHtml(message.message_text)}</div>`;
    chatMessages.appendChild(messageDiv);
    chatMessages.scrollTop = chatMessages.scrollHeight;
    console.log('[CHAT] Message added to chat, scrollTop:', chatMessages.scrollTop);
}

function sendChatMessage() {
    const API_BASE = window.AppConfig.API_BASE;
    const currentUser = window.AppConfig.getCurrentUser();
    
    // Пытаемся получить currentChatId из разных мест
    if (!currentChatId) {
        // Пытаемся загрузить чат для текущего проекта
        const currentProjectId = window.AppConfig.getCurrentProjectId();
        if (currentProjectId) {
            // Пытаемся найти существующий чат
            fetch(`${API_BASE}/chats/projects/${currentProjectId}/chats`, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            })
            .then(res => res.ok ? res.json() : [])
            .then(chats => {
                if (chats && chats.length > 0) {
                    currentChatId = chats[0].id;
                    // Повторяем отправку после установки currentChatId
                    setTimeout(() => sendChatMessage(), 100);
                } else {
                    window.Utils.showMessage('Чат не загружен. Пожалуйста, откройте чат проекта.', 'error');
                }
            })
            .catch(() => {
                window.Utils.showMessage('Чат не загружен. Пожалуйста, откройте чат проекта.', 'error');
            });
            return;
        } else {
            window.Utils.showMessage('Чат не загружен', 'error');
            return;
        }
    }
    
    if (!currentUser) {
        window.Utils.showMessage('Пользователь не авторизован', 'error');
        return;
    }
    
    let input = document.getElementById('chat-message-input');
    if (!input) {
        input = document.querySelector('#project-chat-tab #chat-message-input');
    }
    if (!input) {
        input = document.querySelector('.chat-input-container #chat-message-input');
    }
    if (!input) {
        window.Utils.showMessage('Поле ввода не найдено', 'error');
        return;
    }
    
    const messageText = input.value;
    if (!messageText || messageText.trim().length === 0) {
        window.Utils.showMessage('Введите сообщение', 'error');
        return;
    }
    
    const trimmedText = messageText.trim();
    
    console.log('[CHAT] sendChatMessage - currentChatId:', currentChatId, 'WebSocket state:', currentChatWS ? currentChatWS.readyState : 'null');
    
    if (currentChatWS && currentChatWS.readyState === WebSocket.OPEN) {
        const messageToSend = { user_id: currentUser.id, message_text: trimmedText };
        console.log('[CHAT] Sending message via WebSocket:', messageToSend);
        currentChatWS.send(JSON.stringify(messageToSend));
        input.value = '';
    } else {
        console.log('[CHAT] WebSocket not available, using HTTP fallback');
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
        .then((newMessage) => {
            console.log('[CHAT] Message sent via HTTP, response:', newMessage);
            input.value = '';
            // Добавляем сообщение вручную, так как WebSocket может не сработать
            if (newMessage) {
                addMessageToChat(newMessage);
            } else {
                // Перезагружаем сообщения
                loadChatMessages(currentChatId);
            }
        })
        .catch(err => {
            console.error('[CHAT] Error sending message via HTTP:', err);
            window.Utils.showMessage('Ошибка отправки сообщения: ' + err.message, 'error');
        });
    }
}

function loadChatsPage() {
    const API_BASE = window.AppConfig.API_BASE;
    const selector = document.getElementById('chat-project-select');
    const projectList = document.getElementById('chat-project-list');
    const content = document.getElementById('chats-content');
    
    if (!selector || !content) return;
    
    // Загружаем проекты пользователя (где он участвует - только участники могут видеть чат)
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
        if (projectList) {
            projectList.innerHTML = '';
        }
        
        if (projects && Array.isArray(projects) && projects.length > 0) {
            projects.forEach(project => {
                const option = document.createElement('option');
                option.value = project.id;
                option.textContent = project.title;
                selector.appendChild(option);
                
                // Добавляем проект в список (Telegram-стиль)
                if (projectList) {
                    const projectItem = document.createElement('div');
                    projectItem.className = 'chat-project-item';
                    projectItem.onclick = () => {
                        selector.value = project.id;
                        loadProjectChats();
                    };
                    projectItem.innerHTML = `
                        <div class="chat-project-avatar">${project.title.charAt(0).toUpperCase()}</div>
                        <div class="chat-project-info">
                            <div class="chat-project-name">${window.Utils.escapeHtml(project.title)}</div>
                            <div class="chat-project-meta">${project.role || 'Участник'}</div>
                        </div>
                    `;
                    projectList.appendChild(projectItem);
                }
            });
        } else {
            selector.innerHTML = '<option value="">У вас нет проектов</option>';
            if (projectList) {
                projectList.innerHTML = '<div style="padding: var(--spacing-lg); text-align: center; color: var(--text-secondary);">У вас нет проектов</div>';
            }
        }
    })
    .catch(err => {
        console.error('Error loading projects:', err);
        selector.innerHTML = '<option value="">Ошибка загрузки</option>';
        if (projectList) {
            projectList.innerHTML = '<div style="padding: var(--spacing-lg); text-align: center; color: var(--danger);">Ошибка загрузки проектов</div>';
        }
    });
    
    if (content) {
        content.innerHTML = `
            <div class="chat-empty-state">
                <div class="empty-state-icon">💬</div>
                <h3>Выберите проект для просмотра чата</h3>
                <p>Список ваших проектов отображается слева</p>
            </div>
        `;
    }
}

function loadProjectChats() {
    // Предотвращаем множественные вызовы
    if (isLoadingChats) {
        return;
    }
    
    const API_BASE = window.AppConfig.API_BASE;
    const selector = document.getElementById('chat-project-select');
    const projectId = selector ? selector.value : null;
    const content = document.getElementById('chats-content');
    
    if (!projectId) {
        isLoadingChats = false;
        if (content) {
            content.innerHTML = `
                <div class="chat-empty-state">
                    <div class="empty-state-icon">💬</div>
                    <h3>Выберите проект для просмотра чата</h3>
                    <p>Список ваших проектов отображается слева</p>
                </div>
            `;
        }
        return;
    }
    
    isLoadingChats = true;
    window.AppConfig.setCurrentProjectId(projectId);
    
    // Создаем ТОЧНО такую же структуру, как в проекте - с оберткой project-tab-content
    if (content) {
        content.innerHTML = `
            <div id="standalone-chat-tab" class="project-tab-content" style="display: block;">
                <div class="project-chat-container">
                    <div id="standalone-chat-messages" class="chat-messages-container scrollable-content"></div>
                    <div id="standalone-chat-input-container" class="chat-input-container">
                        <input type="text" id="standalone-chat-message-input" placeholder="Введите сообщение..." onkeypress="if(event.key==='Enter') { event.preventDefault(); window.Chat.sendStandaloneChatMessage(); }">
                        <button onclick="window.Chat.sendStandaloneChatMessage()" class="send-button">
                            <svg width="20" height="20" viewBox="0 0 20 20" fill="none">
                                <path d="M18 2L9 11M18 2L12 18L9 11M18 2L2 8L9 11" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                            </svg>
                        </button>
                    </div>
                </div>
            </div>
        `;
    }
    
    // Используем ТОЧНО такую же логику, как в loadProjectChatMessages
    const chatMessages = document.getElementById('standalone-chat-messages');
    if (!chatMessages) {
        setTimeout(() => loadProjectChats(), 100);
        return;
    }
    
    chatMessages.innerHTML = '<p style="text-align: center; padding: var(--spacing-xl); color: var(--text-secondary);">Загрузка чата...</p>';
    
    fetch(`${API_BASE}/chats/projects/${projectId}/chats`, { 
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` } 
    })
    .then(res => res.ok ? res.json() : [])
    .then(chats => {
        if (!chats || !Array.isArray(chats) || chats.length === 0) { 
            createProjectChat(projectId); 
            return; 
        }
        currentChatId = chats[0].id;
        loadStandaloneChatMessages(chats[0].id);
        connectWebSocket(chats[0].id);
        isLoadingChats = false;
    })
    .catch(err => { 
        console.error('Error loading chat:', err); 
        isLoadingChats = false;
        if (chatMessages) {
            chatMessages.innerHTML = '<p style="text-align: center; padding: var(--spacing-xl); color: var(--danger);">Ошибка загрузки чата</p>';
        }
    });
}

// Отдельная функция для загрузки сообщений в standalone чат - копия loadChatMessages, но для standalone
function loadStandaloneChatMessages(chatId) {
    const API_BASE = window.AppConfig.API_BASE;
    const currentUser = window.AppConfig.getCurrentUser();
    if (!chatId) {
        console.error('loadStandaloneChatMessages: chatId is null');
        return;
    }
    
    const chatMessages = document.getElementById('standalone-chat-messages');
    if (!chatMessages) {
        console.error('standalone-chat-messages container not found');
        return;
    }
    
    fetch(`${API_BASE}/chats/chats/${chatId}/messages`, { 
        headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` } 
    })
    .then(res => {
        if (!res.ok) {
            console.error('Failed to load messages:', res.status);
            return [];
        }
        return res.json();
    })
    .then(messages => {
        if (!messages || !Array.isArray(messages)) messages = [];
        if (messages.length === 0) {
            chatMessages.innerHTML = '<p style="text-align: center; color: var(--text-secondary); padding: var(--spacing-xl);">Нет сообщений</p>';
            return;
        }
        chatMessages.innerHTML = messages.map(msg => `
            <div class="chat-message ${msg.user_id === currentUser.id ? 'own' : ''}">
                <div class="message-header">
                    <strong>${msg.user_id === currentUser.id ? 'Вы' : 'Участник'}</strong>
                    <span class="message-time">${new Date(msg.sent_at).toLocaleString('ru-RU')}</span>
                </div>
                <div class="message-text">${window.Utils.escapeHtml(msg.message_text)}</div>
            </div>
        `).join('');
        chatMessages.scrollTop = chatMessages.scrollHeight;
    })
    .catch(err => { 
        console.error('Error loading messages:', err);
        chatMessages.innerHTML = '<p style="color: var(--danger); text-align: center; padding: var(--spacing-xl);">Ошибка загрузки сообщений</p>';
    });
}

// Отдельная функция для отправки сообщений в standalone чат
function sendStandaloneChatMessage() {
    const API_BASE = window.AppConfig.API_BASE;
    const currentUser = window.AppConfig.getCurrentUser();
    const input = document.getElementById('standalone-chat-message-input');
    const messageText = input ? input.value.trim() : '';
    
    if (!messageText) {
        return;
    }
    
    if (!currentUser) {
        window.Utils.showMessage('Пользователь не авторизован', 'error');
        return;
    }
    
    if (!currentChatId) {
        const currentProjectId = window.AppConfig.getCurrentProjectId();
        if (currentProjectId) {
            fetch(`${API_BASE}/chats/projects/${currentProjectId}/chats`, {
                headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
            })
            .then(res => res.ok ? res.json() : [])
            .then(chats => {
                if (chats && chats.length > 0) {
                    currentChatId = chats[0].id;
                    setTimeout(() => sendStandaloneChatMessage(), 100);
                }
            });
        }
        return;
    }
    
    console.log('[CHAT] sendStandaloneChatMessage - currentChatId:', currentChatId, 'WebSocket state:', currentChatWS ? currentChatWS.readyState : 'null');
    
    // Пробуем отправить через WebSocket
    if (currentChatWS && currentChatWS.readyState === WebSocket.OPEN) {
        const messageToSend = { user_id: currentUser.id, message_text: messageText };
        console.log('[CHAT] Sending standalone message via WebSocket:', messageToSend);
        currentChatWS.send(JSON.stringify(messageToSend));
        if (input) input.value = '';
    } else {
        console.log('[CHAT] WebSocket not available, using HTTP fallback for standalone');
        // Fallback на HTTP если WebSocket не работает
        fetch(`${API_BASE}/chats/chats/${currentChatId}/messages`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${localStorage.getItem('token')}`
            },
            body: JSON.stringify({ message_text: messageText })
        })
        .then(res => {
            if (!res.ok) {
                throw new Error('Failed to send message');
            }
            return res.json();
        })
        .then((newMessage) => {
            console.log('[CHAT] Standalone message sent via HTTP, response:', newMessage);
            if (input) input.value = '';
            // Добавляем сообщение вручную, так как WebSocket может не сработать
            if (newMessage) {
                addMessageToChat(newMessage);
            } else {
                // Перезагружаем сообщения
                loadStandaloneChatMessages(currentChatId);
            }
        })
        .catch(err => {
            console.error('[CHAT] Error sending standalone message:', err);
            window.Utils.showMessage('Ошибка отправки сообщения: ' + err.message, 'error');
        });
    }
}

// Экспорт функций
window.Chat = {
    loadProjectChat,
    loadProjectChatMessages,
    createProjectChat,
    loadChatMessages,
    connectWebSocket,
    addMessageToChat,
    sendChatMessage,
    loadChatsPage,
    loadProjectChats,
    loadStandaloneChatMessages,
    sendStandaloneChatMessage
};

// Глобальные функции для использования в HTML
window.sendChatMessage = sendChatMessage;
window.loadProjectChat = loadProjectChat;
window.loadProjectChats = loadProjectChats;

// Глобальные функции для использования в HTML
window.sendChatMessage = sendChatMessage;
window.loadProjectChat = loadProjectChat;
window.loadProjectChats = loadProjectChats;

