window.InvitationsModule = {
  async loadInvitations() {
    try {
      this.invitations = await api.get('/projects/invitations/my') || [];
      this.invitationsCount = (this.invitations || []).length;
    } catch { this.invitations = []; this.invitationsCount = 0; }
  },
  async refreshInvitationsBadge() {
    try {
      const r = await api.get('/projects/invitations/count');
      this.invitationsCount = (r && r.count) || 0;
    } catch { this.invitationsCount = 0; }
  },

  async acceptInvitation(inv) {
    try {
      const res = await api.post(`/projects/invitations/${inv.id}/accept`, {});
      this.notify('Приглашение принято', 'success');
      await this.loadInvitations();
      if (res && res.project_id) this.openProject(res.project_id);
    } catch (e) { this.notify(e.message, 'error'); }
  },
  async rejectInvitation(inv) {
    if (!confirm('Отклонить приглашение?')) return;
    try {
      await api.post(`/projects/invitations/${inv.id}/reject`, {});
      this.notify('Приглашение отклонено', 'success');
      await this.loadInvitations();
    } catch (e) { this.notify(e.message, 'error'); }
  },

  openInviteForm() {
    this.inviteDraft = { mode: 'email', emails: [], emailInput: '', group_ids: [], groupSearch: '', message: '' };
    this.showInviteForm = true;
    if (!this.groups.length) api.get('/users/groups').then(g => this.groups = g || []);
  },
  setInviteMode(m) {
    this.inviteDraft.mode = m;
    if (m === 'email') { this.inviteDraft.group_ids = []; this.inviteDraft.groupSearch = ''; }
    else { this.inviteDraft.emails = []; this.inviteDraft.emailInput = ''; }
  },
  addInviteEmail() {
    const v = (this.inviteDraft.emailInput || '').trim();
    if (!v) return;
    const ok = /.+@.+\..+/.test(v);
    if (!ok) { this.notify('Некорректный email', 'error'); return; }
    if (!this.inviteDraft.emails.includes(v)) this.inviteDraft.emails.push(v);
    this.inviteDraft.emailInput = '';
  },
  removeInviteEmail(v) {
    this.inviteDraft.emails = this.inviteDraft.emails.filter(e => e !== v);
  },
  toggleInviteGroup(id) {
    if (this.inviteDraft.group_ids.includes(id)) {
      this.inviteDraft.group_ids = this.inviteDraft.group_ids.filter(g => g !== id);
    } else {
      this.inviteDraft.group_ids = [...this.inviteDraft.group_ids, id];
    }
  },
  filteredGroupsForInvite() {
    const q = (this.inviteDraft.groupSearch || '').trim().toLowerCase();
    if (!q) return this.groups;
    return this.groups.filter(g => g.name.toLowerCase().includes(q));
  },
  async submitInvitations() {
    const mode = this.inviteDraft.mode;
    if (mode === 'email' && !this.inviteDraft.emails.length) {
      this.notify('Добавьте хотя бы один email', 'error');
      return;
    }
    if (mode === 'group' && !this.inviteDraft.group_ids.length) {
      this.notify('Выберите хотя бы одну группу', 'error');
      return;
    }
    try {
      const body = { message: this.inviteDraft.message };
      if (mode === 'email') { body.emails = this.inviteDraft.emails; body.group_ids = []; }
      else { body.group_ids = this.inviteDraft.group_ids; body.emails = []; }
      const res = await api.post(`/projects/projects/${this.selectedProject.id}/invitations`, body);
      this.notify(`Приглашений отправлено: ${(res && res.sent) || 0}`, 'success');
      this.showInviteForm = false;
      await this.loadProjectInvitations();
    } catch (e) { this.notify(e.message, 'error'); }
  },
  async loadProjectInvitations() {
    try {
      const all = await api.get(`/projects/projects/${this.selectedProject.id}/invitations`) || [];
      this.projectInvitations = all.filter(i => i.status === 'pending');
    }
    catch { this.projectInvitations = []; }
  },
  async cancelProjectInvitation(inv) {
    if (!confirm('Отменить приглашение?')) return;
    try {
      await api.post(`/projects/invitations/${inv.id}/cancel`, {});
      this.notify('Приглашение отменено', 'success');
      await this.loadProjectInvitations();
    } catch (e) { this.notify(e.message, 'error'); }
  },
  async cancelAllProjectInvitations() {
    const pending = (this.projectInvitations || []).filter(i => i.status === 'pending');
    if (!pending.length) return;
    if (!confirm(`Отменить все приглашения (${pending.length})?`)) return;
    try {
      for (const inv of pending) {
        try { await api.post(`/projects/invitations/${inv.id}/cancel`, {}); } catch {}
      }
      this.notify('Приглашения отменены', 'success');
      await this.loadProjectInvitations();
    } catch (e) { this.notify(e.message, 'error'); }
  },

  invStatusLabel(s) {
    return ({ pending: 'Ожидает', accepted: 'Принято', rejected: 'Отклонено', cancelled: 'Отменено' }[s] || s);
  },
  invStatusClass(s) {
    return ({
      pending:   'bg-amber-100 text-amber-700',
      accepted:  'bg-emerald-100 text-emerald-700',
      rejected:  'bg-rose-100 text-rose-700',
      cancelled: 'bg-ink-100 text-ink-500',
    }[s] || 'bg-ink-100 text-ink-700');
  },
};
