window.AdminModule = {
  setAdminTab(t) {
    this.adminTab = t;
    this.pushUrl('admin', { tab: t });
    this.loadAdminTab();
  },

  async loadAdminTab() {
    try {
      if (this.adminTab === 'dashboard') {
        this.dashboard = await api.get('/users/admin/dashboard');
        try { this.kpiData = await api.get('/reports/kpi/dashboard'); } catch { this.kpiData = null; }
        setTimeout(() => this.renderDashboardCharts(), 150);
      } else if (this.adminTab === 'users') {
        this.adminUsers = await api.get('/users/admin/users') || [];
        if (!this.groups.length) {
          try { this.groups = await api.get('/users/groups') || []; } catch { this.groups = []; }
        }
      } else if (this.adminTab === 'requests') {
        this.adminRequests = await api.get('/users/organizer-requests') || [];
      } else if (this.adminTab === 'projects') {
        let path = '/projects/projects';
        if (this.adminProjectFilter) path += '?status=' + encodeURIComponent(this.adminProjectFilter);
        this.adminProjects = await api.get(path) || [];
      } else if (this.adminTab === 'reports') {
        this.summary = await api.get('/reports/summary');
        await this.loadKpiTables();
      } else if (this.adminTab === 'dictionaries') {
        await this.loadDictionary();
      }
    } catch (e) { this.notify(e.message, 'error'); }
  },

  setDictTab(t) {
    this.dictTab = t;
    this.dictSearch = '';
    this.dictPage = 1;
    this.loadDictionary();
  },
  async loadDictionary() {
    try {
      if (this.dictTab === 'categories') {
        this.dictCategories = await api.get('/projects/categories') || [];
      } else if (this.dictTab === 'tags') {
        this.dictTags = await api.get('/projects/tag-catalog') || [];
      } else if (this.dictTab === 'groups') {
        this.dictGroups = await api.get('/users/groups') || [];
      }
    } catch (e) { this.notify(e.message, 'error'); }
  },
  dictList() {
    if (this.dictTab === 'categories') return this.dictCategories;
    if (this.dictTab === 'tags') return this.dictTags;
    if (this.dictTab === 'groups') return this.dictGroups;
    return [];
  },
  filteredDict() {
    const q = (this.dictSearch || '').trim().toLowerCase();
    const list = this.dictList();
    if (!q) return list;
    return list.filter(x => (x.name || '').toLowerCase().includes(q));
  },
  pagedDict() {
    return paginate(this.filteredDict(), this.dictPage, 50);
  },
  dictPagesCount() {
    return Math.max(1, Math.ceil(this.filteredDict().length / 50));
  },
  openDictForm(item) {
    this.dictDraft = item ? { id: item.id, name: item.name } : { id: null, name: '' };
    this.showDictForm = true;
  },
  async saveDict() {
    if (!this.dictDraft.name.trim()) {
      this.notify('Название обязательно', 'error');
      return;
    }
    try {
      let url;
      if (this.dictTab === 'categories') url = '/projects/admin/categories';
      else if (this.dictTab === 'tags') url = '/projects/admin/tag-catalog';
      else url = '/users/admin/groups';
      if (this.dictDraft.id) {
        await api.put(url + '/' + this.dictDraft.id, { name: this.dictDraft.name });
        this.notify('Запись обновлена', 'success');
      } else {
        await api.post(url, { name: this.dictDraft.name });
        this.notify('Запись создана', 'success');
      }
      this.showDictForm = false;
      this.loadDictionary();
    } catch (e) { this.notify(e.message, 'error'); }
  },
  async deleteDict(item) {
    if (!confirm(`Удалить «${item.name}»?`)) return;
    try {
      let url;
      if (this.dictTab === 'categories') url = '/projects/admin/categories';
      else if (this.dictTab === 'tags') url = '/projects/admin/tag-catalog';
      else url = '/users/admin/groups';
      await api.del(url + '/' + item.id);
      this.notify('Запись удалена', 'success');
      this.loadDictionary();
    } catch (e) { this.notify(e.message, 'error'); }
  },
  dictTabLabel() {
    return ({ categories: 'категория', tags: 'тег', groups: 'группа' }[this.dictTab] || 'запись');
  },

  async loadKpiTables() {
    try {
      const params = new URLSearchParams();
      if (this.reportFrom) params.set('from', this.reportFrom);
      if (this.reportTo) params.set('to', this.reportTo);
      if (this.reportGroupID) params.set('group_id', this.reportGroupID);
      if (this.reportUserType) params.set('user_type', this.reportUserType);
      const qs = params.toString() ? '?' + params.toString() : '';
      const datesOnly = new URLSearchParams();
      if (this.reportFrom) datesOnly.set('from', this.reportFrom);
      if (this.reportTo) datesOnly.set('to', this.reportTo);
      const dQs = datesOnly.toString() ? '?' + datesOnly.toString() : '';
      this.kpiUsers = await api.get('/reports/kpi/users' + qs) || [];
      this.kpiProjects = await api.get('/reports/kpi/projects' + dQs) || [];
      if (!this.groups.length) {
        try { this.groups = await api.get('/users/groups') || []; } catch { this.groups = []; }
      }
    } catch (e) { this.notify(e.message, 'error'); }
  },

  userTypeLabel(t) {
    return ({ student: 'Студент', teacher: 'Преподаватель', staff: 'Сотрудник' }[t] || t);
  },
  userTypeBadgeClass(t) {
    return ({
      student: 'bg-sky-100 text-sky-700',
      teacher: 'bg-violet-100 text-violet-700',
      staff:   'bg-amber-100 text-amber-700',
    }[t] || 'bg-sky-100 text-sky-700');
  },

  async downloadKpiExcel() {
    try {
      const params = new URLSearchParams();
      if (this.reportFrom) params.set('from', this.reportFrom);
      if (this.reportTo) params.set('to', this.reportTo);
      if (this.reportGroupID) params.set('group_id', this.reportGroupID);
      if (this.reportUserType) params.set('user_type', this.reportUserType);
      const qs = params.toString() ? '?' + params.toString() : '';
      const blob = await api.blob('/reports/excel/kpi' + qs);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a'); a.href = url; a.download = 'kpi_report.xlsx';
      document.body.appendChild(a); a.click(); a.remove();
      URL.revokeObjectURL(url);
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async downloadProjectReport(id) {
    try {
      const params = new URLSearchParams();
      if (this.reportFrom) params.set('from', this.reportFrom);
      if (this.reportTo) params.set('to', this.reportTo);
      const qs = params.toString() ? '?' + params.toString() : '';
      const blob = await api.blob('/reports/excel/project-kpi/' + id + qs);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a'); a.href = url; a.download = 'project_kpi.xlsx';
      document.body.appendChild(a); a.click(); a.remove();
      URL.revokeObjectURL(url);
    } catch (e) { this.notify(e.message, 'error'); }
  },

  renderDashboardCharts() {
    if (!this.kpiData || typeof Chart === 'undefined') return;

    const tCanvas = document.getElementById('chart-tasks');
    if (tCanvas) {
      this.destroyChart('tasks');
      const ts = this.kpiData.timeseries || {};
      const labels = (ts.tasks_created || []).map(p => p.date.slice(5));
      this.charts.tasks = new Chart(tCanvas, {
        type: 'line',
        data: { labels,
          datasets: [
            { label: 'Создано', data: (ts.tasks_created||[]).map(p=>p.value), borderColor: '#6366f1', backgroundColor: 'rgba(99,102,241,0.1)', tension: 0.3, fill: true },
            { label: 'Завершено', data: (ts.tasks_completed||[]).map(p=>p.value), borderColor: '#10b981', backgroundColor: 'rgba(16,185,129,0.1)', tension: 0.3, fill: true },
          ] },
        options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { position: 'top' } }, scales: { y: { beginAtZero: true, ticks: { precision: 0 } } } }
      });
    }
    const pCanvas = document.getElementById('chart-project-status');
    if (pCanvas) {
      this.destroyChart('projectStatus');
      const psb = this.kpiData.project_status_breakdown || {};
      this.charts.projectStatus = new Chart(pCanvas, {
        type: 'doughnut',
        data: { labels: ['Активные', 'Завершённые', 'Архив'],
          datasets: [{ data: [psb.active||0, psb.completed||0, psb.archived||0], backgroundColor: ['#10b981','#6366f1','#94a3b8'], borderWidth: 0 }] },
        options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { position: 'bottom' } }, cutout: '65%' }
      });
    }
    const uCanvas = document.getElementById('chart-top-users');
    if (uCanvas) {
      this.destroyChart('topUsers');
      const top = this.kpiData.top_users || [];
      this.charts.topUsers = new Chart(uCanvas, {
        type: 'bar',
        data: { labels: top.map(u => u.first_name + ' ' + u.last_name),
          datasets: [
            { label: 'Задач выполнено', data: top.map(u => u.tasks_completed), backgroundColor: '#6366f1' },
            { label: 'Вклад (событий)', data: top.map(u => u.activity_score), backgroundColor: '#fbbf24' },
          ] },
        options: { responsive: true, maintainAspectRatio: false, indexAxis: 'y', plugins: { legend: { position: 'top' } }, scales: { x: { beginAtZero: true, ticks: { precision: 0 } } } }
      });
    }
  },

  filteredAdminUsers() {
    const q = (this.adminUserSearch || '').trim().toLowerCase();
    if (!q) return this.adminUsers;
    return this.adminUsers.filter(u =>
      (u.email || '').toLowerCase().includes(q) ||
      ((u.first_name || '') + ' ' + (u.last_name || '')).toLowerCase().includes(q) ||
      (u.group_name || '').toLowerCase().includes(q));
  },
  pagedAdminUsers() {
    return paginate(this.filteredAdminUsers(), this.adminUserPage, 50);
  },
  adminUsersPagesCount() {
    return Math.max(1, Math.ceil(this.filteredAdminUsers().length / 50));
  },
  filteredAdminProjects() {
    return this.adminProjects;
  },
  pagedAdminProjects() {
    return paginate(this.filteredAdminProjects(), this.adminProjectPage, 50);
  },
  adminProjectsPagesCount() {
    return Math.max(1, Math.ceil(this.filteredAdminProjects().length / 50));
  },

  openCreateUser() {
    this.userDraft = { id: null, first_name: '', last_name: '', email: '', password: '',
      status: true, is_participant: true, is_organizer: false, is_admin: false,
      user_type: 'student', group_id: '' };
    if (!this.groups.length) api.get('/users/groups').then(g => this.groups = g || []);
    this.showUserForm = true;
  },

  openEditUser(u) {
    this.userDraft = {
      id: u.id, first_name: u.first_name, last_name: u.last_name, email: u.email,
      password: '', status: u.status,
      is_participant: u.is_participant, is_organizer: u.is_organizer, is_admin: u.is_admin,
      user_type: u.user_type || 'student',
      group_id: u.group_id || '',
    };
    if (!this.groups.length) api.get('/users/groups').then(g => this.groups = g || []);
    this.showUserForm = true;
  },

  async saveUser() {
    try {
      const body = { ...this.userDraft };
      delete body.id;
      if (this.userDraft.id) {
        await api.put('/users/admin/users/' + this.userDraft.id, body);
        this.notify('Пользователь обновлён', 'success');
      } else {
        if (!body.password) { this.notify('Пароль обязателен', 'error'); return; }
        await api.post('/users/admin/users', body);
        this.notify('Пользователь создан', 'success');
      }
      this.showUserForm = false;
      this.loadAdminTab();
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async deleteUser(u) {
    if (!confirm(`Удалить пользователя ${u.email}? Действие необратимо.`)) return;
    try {
      await api.del('/users/admin/users/' + u.id);
      this.notify('Пользователь удалён', 'success');
      this.loadAdminTab();
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async toggleUserStatus(u) {
    try {
      const body = {
        first_name: u.first_name, last_name: u.last_name, email: u.email,
        password: '', status: !u.status,
        is_participant: u.is_participant, is_organizer: u.is_organizer, is_admin: u.is_admin,
      };
      await api.put('/users/admin/users/' + u.id, body);
      this.loadAdminTab();
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async deleteProjectAdmin(p) {
    if (!confirm(`Удалить проект «${p.title}»? Все задачи и чаты будут удалены.`)) return;
    try {
      await api.del('/projects/admin/projects/' + p.id);
      this.notify('Проект удалён', 'success');
      this.loadAdminTab();
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async archiveProjectAdmin(p) {
    try { await api.post('/projects/projects/' + p.id + '/archive'); this.loadAdminTab(); this.notify('Проект архивирован', 'success'); }
    catch (e) { this.notify(e.message, 'error'); }
  },

  async restoreProjectAdmin(p) {
    try { await api.post('/projects/admin/projects/' + p.id + '/restore'); this.loadAdminTab(); this.notify('Проект восстановлен', 'success'); }
    catch (e) { this.notify(e.message, 'error'); }
  },

  async approveOrganizer(id) {
    try { await api.post('/users/organizer-requests/' + id + '/approve'); this.notify('Заявка одобрена', 'success'); this.loadAdminTab(); }
    catch (e) { this.notify(e.message, 'error'); }
  },

  async rejectOrganizer(id) {
    try { await api.post('/users/organizer-requests/' + id + '/reject'); this.notify('Заявка отклонена', 'success'); this.loadAdminTab(); }
    catch (e) { this.notify(e.message, 'error'); }
  },

  filteredAdminRequests() {
    const f = this.adminRequestFilter || 'all';
    const tf = this.adminRequestTypeFilter || 'all';
    const list = (this.adminRequests || []).filter(r => {
      if (tf !== 'all' && (r.request_type || 'organizer') !== tf) return false;
      if (f === 'all') return true;
      if (f === 'pending') return r.status === 'в рассмотрении';
      if (f === 'approved') return r.status === 'одобрена';
      if (f === 'rejected') return r.status === 'отклонена';
      return true;
    });
    return list.slice().sort((a, b) => {
      const ap = a.status === 'в рассмотрении' ? 0 : 1;
      const bp = b.status === 'в рассмотрении' ? 0 : 1;
      if (ap !== bp) return ap - bp;
      return new Date(b.submission_date) - new Date(a.submission_date);
    });
  },

  adminRequestCount(status) {
    const tf = this.adminRequestTypeFilter || 'all';
    return (this.adminRequests || []).filter(r => {
      if (tf !== 'all' && (r.request_type || 'organizer') !== tf) return false;
      if (status === 'all') return true;
      if (status === 'pending') return r.status === 'в рассмотрении';
      if (status === 'approved') return r.status === 'одобрена';
      if (status === 'rejected') return r.status === 'отклонена';
      return false;
    }).length;
  },
};
