window.AuthModule = {
  init() {
    this.setupRouter();
    const token = localStorage.getItem('token');
    if (token) {
      api.get('/users/profile').then(u => {
        this.user = u;
        this.role = localStorage.getItem('role');
        if (this.role !== 'admin') this.refreshInvitationsBadge();
        const r = this.parseUrl();
        if (r.page === 'project-detail' && r.id) {
          this.openProject(r.id);
        } else if (r.page === 'admin' || this.role === 'admin') {
          this.page = 'admin';
          this.adminTab = r.tab || 'dashboard';
          this.pushUrl('admin', { tab: this.adminTab });
          this.loadAdminTab();
        } else if (['projects','my-projects','chats','profile','notifications'].includes(r.page)) {
          this.go(r.page);
        } else {
          this.go('projects');
        }
      }).catch(() => this.logout());
    } else {
      const r = this.parseUrl();
      this.page = r.page === 'register' ? 'register' : 'home';
      this.pushUrl(this.page);
    }
  },

  notify(text, type = 'info') {
    this.toast = { show: true, type, text };
    setTimeout(() => this.toast.show = false, 3000);
  },

  go(page, opts) {
    opts = opts || {};
    if (!this.user && page !== 'home' && page !== 'register') page = 'home';
    if (this.role === 'admin' && page !== 'admin' && page !== 'profile' && page !== 'project-detail') page = 'admin';
    this.page = page;
    this.selectedProject = null;
    this.selectedTask = null;
    this.showProjectForm = false;
    if (!opts.fromRouter) this.pushUrl(page, page === 'admin' ? { tab: this.adminTab } : {});
    if (page === 'projects') this.loadProjects();
    else if (page === 'my-projects') this.loadMyProjects();
    else if (page === 'chats') this.loadChatsPage();
    else if (page === 'notifications') this.loadInvitations();
    else if (page === 'admin') this.loadAdminTab();
  },

  async login() {
    try {
      const data = await api.post('/users/login', this.loginForm);
      localStorage.setItem('token', data.token);
      localStorage.setItem('role', data.role);
      this.user = data.user;
      this.role = data.role;
      this.loginForm = { email: '', password: '' };
      if (this.role !== 'admin') this.refreshInvitationsBadge();
      this.go(this.role === 'admin' ? 'admin' : 'projects');
      this.notify('Добро пожаловать, ' + data.user.first_name + '!', 'success');
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async register() {
    if (!this.acceptPdn) { this.notify('Подтвердите согласие на обработку персональных данных', 'error'); return; }
    try {
      await api.post('/users/register', this.registerForm);
      this.notify('Регистрация выполнена, войдите в систему', 'success');
      this.registerForm = { first_name: '', last_name: '', email: '', password: '' };
      this.go('home');
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async uploadField(ev, draftName, fieldName) {
    const file = ev.target.files && ev.target.files[0];
    console.log('[upload] selected', { draftName, fieldName, file });
    if (!file) return;
    try {
      const res = await api.upload(file);
      console.log('[upload] response', res);
      if (!res || !res.url) throw new Error('Сервер не вернул URL файла');
      if (!this[draftName]) this[draftName] = {};
      this[draftName][fieldName] = res.url;
      console.log('[upload] set', draftName + '.' + fieldName, '=', res.url);
      this.notify('Файл прикреплён: ' + (window.fileLabel(res.url) || file.name || ''), 'success');
    } catch (e) {
      console.error('[upload] failed', e);
      this.notify('Ошибка загрузки: ' + (e.message || 'Unknown'), 'error');
    } finally {
      try { ev.target.value = ''; } catch {}
    }
  },

  logout() {
    localStorage.removeItem('token');
    localStorage.removeItem('role');
    this.user = null;
    this.role = null;
    if (this.ws) { try { this.ws.close(); } catch {} this.ws = null; }
    this.page = 'home';
    if (window.location.pathname !== '/login') {
      window.history.pushState({}, '', '/login');
    }
  },
};
