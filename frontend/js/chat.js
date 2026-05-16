window.ChatModule = {
  async openProjectChat(projectId) {
    try {
      let chats = await api.get(`/chats/projects/${projectId}/chats`) || [];
      if (chats.length === 0) {
        await api.post(`/chats/projects/${projectId}/chats`, { name: 'Общий чат' });
        chats = await api.get(`/chats/projects/${projectId}/chats`) || [];
      }
      this.chats = chats;
      if (chats.length) this.selectChat(chats[0]);
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async selectChat(chat) {
    if (this.ws) { try { this.ws.close(); } catch {} }
    this.selectedChat = chat;
    try { this.messages = await api.get(`/chats/chats/${chat.id}/messages`) || []; }
    catch (e) { this.notify(e.message, 'error'); }
    this.connectWS(chat.id);
    this.$nextTick(() => this.scrollChat());
  },

  connectWS(chatId) {
    const proto = location.protocol === 'https:' ? 'wss' : 'ws';
    this.ws = new WebSocket(`${proto}://${location.host}/api/chats/chats/${chatId}/ws`);
    this.ws.onmessage = (ev) => {
      try {
        const msg = JSON.parse(ev.data);
        if (!this.messages.find(m => m.id === msg.id)) this.messages.push(msg);
        this.$nextTick(() => this.scrollChat());
      } catch {}
    };
  },

  sendMessage() {
    if (!this.newMessage.trim() || !this.selectedChat) return;
    const text = this.newMessage; this.newMessage = '';
    if (this.ws && this.ws.readyState === 1) {
      this.ws.send(JSON.stringify({ user_id: this.user.id, message_text: text }));
    } else {
      api.post(`/chats/chats/${this.selectedChat.id}/messages`, { message_text: text })
        .then(() => api.get(`/chats/chats/${this.selectedChat.id}/messages`))
        .then(msgs => { this.messages = msgs || []; this.$nextTick(() => this.scrollChat()); })
        .catch(e => this.notify(e.message, 'error'));
    }
  },

  scrollChat() {
    const el = document.getElementById('chat-scroll');
    if (el) el.scrollTop = el.scrollHeight;
  },

  async loadChatsPage() {
    try {
      this.chatProjects = await api.get('/projects/projects/my') || [];
      this.activeChatProject = null;
      this.chats = []; this.selectedChat = null; this.messages = [];
    } catch (e) { this.notify(e.message, 'error'); }
  },

  selectChatProject(p) { this.activeChatProject = p; this.openProjectChat(p.id); },
};
