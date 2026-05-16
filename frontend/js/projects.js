window.ProjectsModule = {
  async loadProjects() {
    try {
      let path = '/projects/projects';
      if (this.projectFilter) path += '?status=' + encodeURIComponent(this.projectFilter);
      this.projects = await api.get(path) || [];
      if (!this.categories.length) this.categories = await api.get('/projects/categories') || [];
      if (!this.tagCatalog.length) {
        try { this.tagCatalog = await api.get('/projects/tag-catalog') || []; } catch { this.tagCatalog = []; }
      }
      if (this.user) {
        try { this.recommended = await api.get('/projects/projects/recommended') || []; }
        catch { this.recommended = []; }
      }
    } catch (e) { this.notify(e.message, 'error'); }
  },

  filteredCatalog() {
    const q = (this.themeSearch || '').trim().toLowerCase();
    if (!q) return this.tagCatalog;
    return this.tagCatalog.filter(t => t.name.toLowerCase().includes(q));
  },
  toggleDraftTag(name) {
    if (this.projectDraft.tags.includes(name)) {
      this.projectDraft.tags = this.projectDraft.tags.filter(t => t !== name);
    } else {
      this.projectDraft.tags.push(name);
    }
  },
  toggleDraftCategory(id) {
    const arr = this.projectDraft.category_ids || [];
    if (arr.includes(id)) this.projectDraft.category_ids = arr.filter(x => x !== id);
    else this.projectDraft.category_ids = [...arr, id];
  },

  filteredProjects() {
    const q = (this.projectSearch || '').trim().toLowerCase();
    const cat = this.projectCategoryFilter;
    const list = (this.projects || []).filter(p => {
      if (cat) {
        const ids = p.category_ids && p.category_ids.length ? p.category_ids : (p.category_id ? [p.category_id] : []);
        if (!ids.includes(cat)) return false;
      }
      if (!q) return true;
      return (p.title || '').toLowerCase().includes(q) ||
             (p.short_description || '').toLowerCase().includes(q) ||
             (p.tags || []).some(t => (t || '').toLowerCase().includes(q));
    }).slice();
    const ts = (d) => d ? new Date(d).getTime() : 0;
    const tsMax = Number.MAX_SAFE_INTEGER;
    const sort = this.projectSort;
    if (sort === 'old') list.sort((a, b) => ts(a.creation_date) - ts(b.creation_date));
    else if (sort === 'popular') list.sort((a, b) => (Number(b.participants_count) || 0) - (Number(a.participants_count) || 0));
    else if (sort === 'deadline') list.sort((a, b) => (a.planned_end_date ? ts(a.planned_end_date) : tsMax) - (b.planned_end_date ? ts(b.planned_end_date) : tsMax));
    else if (sort === 'title') list.sort((a, b) => (a.title || '').localeCompare(b.title || '', 'ru'));
    else list.sort((a, b) => ts(b.creation_date) - ts(a.creation_date));
    return list;
  },

  async loadMyProjects() {
    try { this.myProjects = await api.get('/projects/projects/my') || []; }
    catch (e) { this.notify(e.message, 'error'); }
  },

  async openProject(id) {
    try {
      const proj = await api.get('/projects/projects/' + id);
      try { proj.tags = await api.get(`/projects/projects/${id}/tags`) || []; } catch { proj.tags = []; }
      this.selectedProject = proj;
      try { this.requiredSkills = await api.get(`/projects/projects/${id}/required-skills`) || []; } catch { this.requiredSkills = []; }
      if (!this.tagCatalog.length) {
        try { this.tagCatalog = await api.get('/projects/tag-catalog') || []; } catch {}
      }
      await this.computeRights(proj);
      this.projectTab = this.isMember ? 'tasks' : 'about';
      this.myParticipationRequest = null;
      if (this.user && !this.isMember && proj.organizer_id !== this.user.id) {
        try { this.myParticipationRequest = await api.get(`/projects/projects/${id}/my-request`); } catch {}
      }
      this.page = 'project-detail';
      this.pushUrl('project-detail', { id });
      this.loadGoals();
      if (this.isMember) this.loadTasks();
    } catch (e) {
      console.error('openProject error', e);
      this.notify('Не удалось открыть проект: ' + e.message, 'error');
    }
  },

  async computeRights(p) {
    this.canManage = false;
    this.canLead = false;
    this.isMember = false;
    this.participants = [];
    if (!this.user) return;
    try {
      const participants = await api.get(`/projects/projects/${p.id}/participants`) || [];
      this.participants = participants;
      const me = participants.find(x => x.user_id === this.user.id);
      if (this.role === 'admin') {
        this.canManage = true; this.canLead = true; this.isMember = true;
      } else if (this.user.id === p.organizer_id) {
        this.canManage = true; this.canLead = true; this.isMember = true;
      } else if (me) {
        this.isMember = true;
        this.canManage = me.role === 'руководитель' || me.role === 'заместитель';
        this.canLead = me.role === 'руководитель';
      }
    } catch {}
  },

  openCreateProject() {
    this.projectDraft = { id: null, title: '', short_description: '', full_description: '', category_ids: [], tags: [], goal_description: '', planned_end_date: '', image_url: '' };
    this.themeSearch = '';
    this.showProjectForm = true;
    if (!this.categories.length) api.get('/projects/categories').then(c => this.categories = c || []);
    if (!this.tagCatalog.length) api.get('/projects/tag-catalog').then(t => this.tagCatalog = t || []);
  },

  async openEditProject() {
    const p = this.selectedProject;
    this.projectDraft = {
      id: p.id, title: p.title, short_description: p.short_description,
      full_description: p.full_description,
      category_ids: [...(p.category_ids || (p.category_id ? [p.category_id] : []))],
      goal_description: p.goal_description || '',
      planned_end_date: p.planned_end_date ? p.planned_end_date.slice(0,10) : '',
      image_url: p.image_url || '',
      tags: [...(p.tags || [])],
    };
    this.themeSearch = '';
    this.requiredSkillSearch = '';
    if (!this.tagCatalog.length) {
      try { this.tagCatalog = await api.get('/projects/tag-catalog') || []; } catch {}
    }
    if (!this.categories.length) {
      try { this.categories = await api.get('/projects/categories') || []; } catch {}
    }
    try { this.requiredSkills = await api.get(`/projects/projects/${p.id}/required-skills`) || []; } catch {}
    this.showProjectForm = true;
  },

  async saveProject() {
    try {
      const body = {
        title: this.projectDraft.title,
        short_description: this.projectDraft.short_description,
        full_description: this.projectDraft.full_description,
        goal_description: this.projectDraft.goal_description || null,
        planned_end_date: this.projectDraft.planned_end_date || null,
        category_id: (this.projectDraft.category_ids || [])[0] || null,
        category_ids: this.projectDraft.category_ids || [],
        image_url: this.projectDraft.image_url || null,
      };
      let id;
      if (this.projectDraft.id) {
        await api.put('/projects/projects/' + this.projectDraft.id, body);
        id = this.projectDraft.id;
        const current = await api.get(`/projects/projects/${id}/tags`) || [];
        for (const t of current) if (!this.projectDraft.tags.includes(t)) await api.del(`/projects/projects/${id}/tags/${encodeURIComponent(t)}`);
        for (const t of this.projectDraft.tags) if (!current.includes(t)) await api.post(`/projects/projects/${id}/tags`, { name: t });
        this.notify('Проект обновлён', 'success');
      } else {
        const res = await api.post('/projects/projects', body);
        id = res.id;
        for (const t of this.projectDraft.tags) await api.post(`/projects/projects/${id}/tags`, { name: t });
        this.notify('Проект создан', 'success');
      }
      this.showProjectForm = false;
      this.openProject(id);
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async archiveProject() {
    if (!confirm('Архивировать проект?')) return;
    try {
      await api.post(`/projects/projects/${this.selectedProject.id}/archive`);
      this.notify('Проект архивирован', 'success');
      this.go('projects');
    } catch (e) { this.notify(e.message, 'error'); }
  },

  canCompleteProject() {
    if (!this.canLead) return false;
    if (!this.selectedProject || this.selectedProject.status !== 'активен') return false;
    return true;
  },
  allTasksDone() {
    if (!this.tasks.length) return false;
    return this.tasks.every(t => t.status === 'завершена');
  },
  completeBlockReason() {
    if (!this.tasks.length) return 'Сначала создайте и завершите задачи';
    const open = this.tasks.filter(t => t.status !== 'завершена').length;
    if (open) return `Незавершённых задач: ${open}`;
    return '';
  },

  async completeProject() {
    if (!confirm('Завершить проект? После этого он закрывается для редактирования.')) return;
    try {
      await api.post(`/projects/projects/${this.selectedProject.id}/complete`);
      this.notify('Проект завершён', 'success');
      await this.openProject(this.selectedProject.id);
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async submitParticipation() {
    if (this.participateDraft.resume_url && !window.isValidUrl(this.participateDraft.resume_url)) {
      this.notify('Прикрепите резюме (файл или валидная ссылка https://)', 'error');
      return;
    }
    try {
      await api.post(`/projects/projects/${this.selectedProject.id}/participate`, this.participateDraft);
      this.notify('Заявка отправлена', 'success');
      this.showParticipateForm = false;
      this.participateDraft = { comment: '', resume_url: '' };
      this.myParticipationRequest = await api.get(`/projects/projects/${this.selectedProject.id}/my-request`);
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async loadParticipants() {
    try { this.participants = await api.get(`/projects/projects/${this.selectedProject.id}/participants`) || []; }
    catch (e) { this.notify(e.message, 'error'); }
  },

  async loadRequests() {
    try { this.requests = await api.get(`/projects/projects/${this.selectedProject.id}/requests`) || []; }
    catch (e) { this.notify(e.message, 'error'); }
  },

  async approveRequest(id) {
    try { await api.post(`/projects/requests/${id}/approve`); this.loadRequests(); this.loadParticipants(); this.notify('Заявка одобрена', 'success'); }
    catch (e) { this.notify(e.message, 'error'); }
  },

  async rejectRequest(id) {
    try { await api.post(`/projects/requests/${id}/reject`); this.loadRequests(); this.notify('Заявка отклонена', 'success'); }
    catch (e) { this.notify(e.message, 'error'); }
  },

  async setParticipantRole(p, role) {
    try {
      await api.put(`/projects/projects/${this.selectedProject.id}/participants/${p.user_id}/role`, { role });
      this.loadParticipants();
    } catch (e) { this.notify(e.message, 'error'); }
  },

  onProjectTab(t) {
    this.projectTab = t;
    if (t === 'tasks') this.loadTasks();
    else if (t === 'participants') {
      this.loadParticipants();
      if (this.canManage) { this.loadRequests(); this.loadProjectInvitations(); }
    }
    else if (t === 'chat') this.openProjectChat(this.selectedProject.id);
    else if (t === 'goals') this.loadGoals();
    else if (t === 'analytics') this.loadProjectAnalytics();
  },

  async loadProjectAnalytics() {
    try {
      const params = new URLSearchParams();
      if (this.projectAnalyticsFrom) params.set('from', this.projectAnalyticsFrom);
      if (this.projectAnalyticsTo) params.set('to', this.projectAnalyticsTo);
      const qs = params.toString() ? '?' + params.toString() : '';
      this.projectDashboard = await api.get(`/reports/projects/${this.selectedProject.id}/dashboard${qs}`);
      setTimeout(() => this.renderProjectCharts(), 150);
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async downloadProjectKPIExcel() {
    try {
      const params = new URLSearchParams();
      if (this.projectAnalyticsFrom) params.set('from', this.projectAnalyticsFrom);
      if (this.projectAnalyticsTo) params.set('to', this.projectAnalyticsTo);
      const qs = params.toString() ? '?' + params.toString() : '';
      const blob = await api.blob(`/reports/excel/project-kpi/${this.selectedProject.id}${qs}`);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a'); a.href = url;
      a.download = 'project_kpi_' + this.selectedProject.id.slice(0,8) + '.xlsx';
      document.body.appendChild(a); a.click(); a.remove();
      URL.revokeObjectURL(url);
    } catch (e) { this.notify(e.message, 'error'); }
  },

  renderProjectCharts() {
    if (!this.projectDashboard || typeof Chart === 'undefined') return;
    const sb = this.projectDashboard.status_breakdown || {};
    const sCanvas = document.getElementById('chart-proj-status');
    if (sCanvas) {
      this.destroyChart('projStatus');
      this.charts.projStatus = new Chart(sCanvas, {
        type: 'doughnut',
        data: { labels: ['Новые', 'В работе', 'Завершены'],
          datasets: [{ data: [sb.new||0, sb.in_progress||0, sb.done||0], backgroundColor: ['#fbbf24','#3b82f6','#10b981'], borderWidth: 0 }] },
        options: { responsive: true, maintainAspectRatio: false, plugins: { legend: { position: 'bottom' } }, cutout: '65%' }
      });
    }
    const uCanvas = document.getElementById('chart-proj-users');
    if (uCanvas) {
      this.destroyChart('projUsers');
      const top = (this.projectDashboard.user_kpis || []).slice(0, 8);
      this.charts.projUsers = new Chart(uCanvas, {
        type: 'bar',
        data: { labels: top.map(u => u.first_name + ' ' + u.last_name),
          datasets: [
            { label: 'Выполнено', data: top.map(u => u.tasks_completed), backgroundColor: '#10b981' },
            { label: 'В срок', data: top.map(u => u.tasks_on_time), backgroundColor: '#6366f1' },
          ] },
        options: { responsive: true, maintainAspectRatio: false, indexAxis: 'y', plugins: { legend: { position: 'top' } }, scales: { x: { beginAtZero: true, ticks: { precision: 0 } } } }
      });
    }
  },

  async loadGoals() {
    try { this.goals = await api.get(`/projects/projects/${this.selectedProject.id}/goals`) || []; }
    catch { this.goals = []; }
  },
  async saveGoal() {
    try {
      await api.post(`/projects/projects/${this.selectedProject.id}/goals`, this.goalDraft);
      this.showGoalForm = false;
      this.goalDraft = { title: '', description: '' };
      this.loadGoals();
      this.notify('Цель добавлена', 'success');
    } catch (e) { this.notify(e.message, 'error'); }
  },
  async toggleGoal(g) {
    try { await api.put(`/projects/goals/${g.id}`, { is_achieved: !g.is_achieved }); this.loadGoals(); }
    catch (e) { this.notify(e.message, 'error'); }
  },
  async deleteGoal(g) {
    if (!confirm('Удалить цель?')) return;
    try { await api.del(`/projects/goals/${g.id}`); this.loadGoals(); }
    catch (e) { this.notify(e.message, 'error'); }
  },
  goalProgress() {
    if (!this.goals.length) return 0;
    return Math.round(this.goals.filter(g => g.is_achieved).length * 100 / this.goals.length);
  },

  filteredRequiredSkillCatalog() {
    const q = (this.requiredSkillSearch || '').trim().toLowerCase();
    const used = new Set(this.requiredSkills || []);
    return (this.tagCatalog || []).filter(t => !used.has(t.name) && (!q || t.name.toLowerCase().includes(q)));
  },
  async addRequiredSkillFromCatalog(name) {
    if (!name) return;
    try {
      await api.post(`/projects/projects/${this.selectedProject.id}/required-skills`, { name });
      this.requiredSkills = await api.get(`/projects/projects/${this.selectedProject.id}/required-skills`) || [];
    } catch (e) { this.notify(e.message, 'error'); }
  },
  async deleteRequiredSkill(name) {
    try {
      await api.del(`/projects/projects/${this.selectedProject.id}/required-skills/${encodeURIComponent(name)}`);
      this.requiredSkills = await api.get(`/projects/projects/${this.selectedProject.id}/required-skills`) || [];
    } catch (e) { this.notify(e.message, 'error'); }
  },

  daysLeft(date) {
    if (!date) return null;
    return Math.ceil((new Date(date) - new Date()) / 86400000);
  },

  myProjectRoleLabel() {
    if (!this.selectedProject || !this.user) return '';
    if (this.role === 'admin') return 'Администратор';
    if (this.user.id === this.selectedProject.organizer_id) return 'Организатор · Руководитель';
    const me = this.participants.find(p => p.user_id === this.user.id);
    if (!me) return 'Не участник';
    return me.role.charAt(0).toUpperCase() + me.role.slice(1);
  },

  destroyChart(name) {
    if (this.charts[name]) { try { this.charts[name].destroy(); } catch {} this.charts[name] = null; }
  },
};
