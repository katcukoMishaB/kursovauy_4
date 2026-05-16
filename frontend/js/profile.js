window.ProfileModule = {
  async loadProfile() {
    try {
      this.user = await api.get('/users/profile');
      await this.loadProfileExtras();
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async loadProfileExtras() {
    try {
      this.mySkills = await api.get('/users/skills') || [];
      this.myInterests = await api.get('/users/interests') || [];
      if (!this.categories.length) this.categories = await api.get('/projects/categories') || [];
      if (!this.tagCatalog.length) this.tagCatalog = await api.get('/projects/tag-catalog') || [];
      if (!this.groups.length) this.groups = await api.get('/users/groups') || [];
      try { this.myOrganizerRequest = await api.get('/users/my-request?type=organizer'); }
      catch { this.myOrganizerRequest = null; }
      try { this.myTeacherRequest = await api.get('/users/my-request?type=teacher'); }
      catch { this.myTeacherRequest = null; }
    } catch (e) { this.notify(e.message, 'error'); }
  },

  startEditProfile() {
    this.profileDraft = {
      first_name: this.user.first_name,
      last_name: this.user.last_name,
      password: '',
      group_id: this.user.group_id || '',
    };
    this.editingProfile = true;
  },

  async saveProfile() {
    try {
      await api.put('/users/profile', this.profileDraft);
      this.editingProfile = false;
      await this.loadProfile();
      this.notify('Профиль обновлён', 'success');
    } catch (e) { this.notify(e.message, 'error'); }
  },

  organizerRequestStatusLabel() {
    const r = this.myOrganizerRequest;
    if (!r || !r.status) return null;
    return r.status;
  },
  teacherRequestStatusLabel() {
    const r = this.myTeacherRequest;
    if (!r || !r.status) return null;
    return r.status;
  },
  hasPendingOrganizerRequest() {
    return this.myOrganizerRequest && this.myOrganizerRequest.status === 'в рассмотрении';
  },
  hasPendingTeacherRequest() {
    return this.myTeacherRequest && this.myTeacherRequest.status === 'в рассмотрении';
  },

  async submitOrganizerRequest() {
    try {
      const draft = this.organizerDraft;
      if (!draft.experience_description || draft.experience_description.trim().length < 10) {
        this.notify('Опишите опыт подробнее (минимум 10 символов)', 'error');
        return;
      }
      if (!draft.resume_url || !window.isValidUrl(draft.resume_url)) {
        this.notify('Прикрепите резюме (файл или валидная ссылка https://)', 'error');
        return;
      }
      await api.post('/users/organizer-request', {
        experience_description: draft.experience_description,
        resume_url: draft.resume_url,
        request_type: 'organizer',
      });
      this.notify('Заявка подана', 'success');
      this.organizerDraft = { experience_description: '', resume_url: '' };
      this.showOrganizerForm = false;
      try { this.myOrganizerRequest = await api.get('/users/my-request?type=organizer'); } catch {}
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async submitTeacherRequest() {
    try {
      const draft = this.teacherDraft;
      if (!draft.experience_description || draft.experience_description.trim().length < 10) {
        this.notify('Опишите опыт подробнее (минимум 10 символов)', 'error');
        return;
      }
      if (!draft.resume_url || !window.isValidUrl(draft.resume_url)) {
        this.notify('Прикрепите подтверждение (файл или валидная ссылка https://)', 'error');
        return;
      }
      await api.post('/users/organizer-request', {
        experience_description: draft.experience_description,
        resume_url: draft.resume_url,
        request_type: 'teacher',
      });
      this.notify('Заявка подана', 'success');
      this.teacherDraft = { experience_description: '', resume_url: '' };
      this.showTeacherForm = false;
      try { this.myTeacherRequest = await api.get('/users/my-request?type=teacher'); } catch {}
    } catch (e) { this.notify(e.message, 'error'); }
  },

  userTypeLabel(t) {
    return ({ student: 'Студент', teacher: 'Преподаватель', staff: 'Сотрудник' }[t] || 'Студент');
  },
  userTypeBadgeClass(t) {
    return ({
      student: 'bg-sky-100 text-sky-700',
      teacher: 'bg-violet-100 text-violet-700',
      staff:   'bg-amber-100 text-amber-700',
    }[t] || 'bg-sky-100 text-sky-700');
  },

  filteredSkillCatalog() {
    const q = (this.skillSearch || '').trim().toLowerCase();
    const used = new Set(this.mySkills || []);
    return this.tagCatalog.filter(t => !used.has(t.name) && (!q || t.name.toLowerCase().includes(q)));
  },

  async addMySkill() {
    const name = (this.skillInput || '').trim();
    if (!name) return;
    try {
      await api.post('/users/skills', { name });
      this.skillInput = '';
      this.mySkills = await api.get('/users/skills') || [];
    } catch (e) { this.notify(e.message, 'error'); }
  },
  async removeMySkill(name) {
    try {
      await api.del('/users/skills/' + encodeURIComponent(name));
      this.mySkills = await api.get('/users/skills') || [];
    } catch (e) { this.notify(e.message, 'error'); }
  },
  async toggleInterest(cat) {
    const has = this.myInterests.find(c => c.category_id === cat.id);
    try {
      if (has) await api.del('/users/interests/' + cat.id);
      else await api.post('/users/interests', { category_id: cat.id });
      this.myInterests = await api.get('/users/interests') || [];
    } catch (e) { this.notify(e.message, 'error'); }
  },
  isInterested(catId) { return !!this.myInterests.find(c => c.category_id === catId); },
};
