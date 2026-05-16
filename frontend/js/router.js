window.RouterModule = {
  parseUrl() {
    const parts = window.location.pathname.split('/').filter(Boolean);
    if (parts.length === 0) return { page: 'projects' };
    const p0 = decodeURIComponent(parts[0]);
    const p1 = parts[1] ? decodeURIComponent(parts[1]) : '';
    if (p0 === 'home' || p0 === 'login') return { page: 'home' };
    if (p0 === 'register') return { page: 'register' };
    if (p0 === 'projects' && p1) return { page: 'project-detail', id: p1 };
    if (p0 === 'projects') return { page: 'projects' };
    if (p0 === 'my-projects') return { page: 'my-projects' };
    if (p0 === 'chats') return { page: 'chats' };
    if (p0 === 'profile') return { page: 'profile' };
    if (p0 === 'admin') return { page: 'admin', tab: p1 || 'dashboard' };
    return { page: 'projects' };
  },

  pageToUrl(page, params) {
    params = params || {};
    if (page === 'home') return '/login';
    if (page === 'register') return '/register';
    if (page === 'projects') return '/projects';
    if (page === 'project-detail') return '/projects/' + (params.id || '');
    if (page === 'my-projects') return '/my-projects';
    if (page === 'chats') return '/chats';
    if (page === 'profile') return '/profile';
    if (page === 'admin') return '/admin' + (params.tab && params.tab !== 'dashboard' ? '/' + params.tab : '');
    return '/';
  },

  pushUrl(page, params) {
    const url = this.pageToUrl(page, params);
    try {
      if (window.location.pathname !== url) {
        window.history.pushState({ page, params }, '', url);
      } else {
        window.history.replaceState({ page, params }, '', url);
      }
    } catch (e) {
      console.error('pushUrl failed', url, e);
    }
  },

  setupRouter() {
    window.addEventListener('popstate', () => this.syncFromUrl());
  },

  syncFromUrl() {
    const r = this.parseUrl();
    if (r.page === 'project-detail' && r.id) {
      this.openProject(r.id);
      return;
    }
    if (r.page === 'admin') {
      this.page = 'admin';
      this.adminTab = r.tab || 'dashboard';
      this.loadAdminTab();
      return;
    }
    if (this.page !== r.page) this.go(r.page, { fromRouter: true });
  },
};
