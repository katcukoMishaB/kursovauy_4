window.TasksModule = {
  async loadTasks() {
    try { this.tasks = await api.get(`/tasks/projects/${this.selectedProject.id}/tasks`) || []; }
    catch (e) { this.notify(e.message, 'error'); }
  },

  tasksByStatus(status) { return this.tasks.filter(t => t.status === status); },

  openCreateTask() {
    this.taskDraft = { id: null, title: '', description: '', priority: 'средний', difficulty: 3, due_date: '', attachment_url: '' };
    this.showTaskForm = true;
  },

  openEditTask(task) {
    this.taskDraft = {
      id: task.id, title: task.title, description: task.description || '',
      priority: task.priority || 'средний', difficulty: task.difficulty || 3,
      due_date: task.due_date ? task.due_date.slice(0,10) : '',
      attachment_url: task.attachment_url || '',
    };
    this.showTaskForm = true;
  },

  async saveTask() {
    try {
      const body = {
        title: this.taskDraft.title,
        description: this.taskDraft.description,
        priority: this.taskDraft.priority,
        difficulty: parseInt(this.taskDraft.difficulty) || 3,
        due_date: this.taskDraft.due_date || null,
        attachment_url: this.taskDraft.attachment_url || null,
      };
      if (this.taskDraft.id) {
        await api.put(`/tasks/tasks/${this.taskDraft.id}`, body);
      } else {
        await api.post(`/tasks/projects/${this.selectedProject.id}/tasks`, body);
      }
      this.showTaskForm = false;
      this.loadTasks();
      this.notify('Задача сохранена', 'success');
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async assignTask(task, userId) {
    try {
      await api.post(`/tasks/tasks/${task.id}/assign`, { assigned_to: userId || null });
      this.loadTasks();
      if (this.selectedTask?.id === task.id) this.selectedTask = (await api.get('/tasks/tasks/' + task.id));
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async setTaskStatus(task, status) {
    try {
      await api.put(`/tasks/tasks/${task.id}/status`, { status });
      this.loadTasks();
      if (this.selectedTask?.id === task.id) this.selectedTask = (await api.get('/tasks/tasks/' + task.id));
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async rateTask(task) {
    if (!this.rateDraft) return;
    try {
      await api.post(`/tasks/tasks/${task.id}/rate`, { quality_rating: this.rateDraft });
      this.loadTasks();
      this.selectedTask = await api.get('/tasks/tasks/' + task.id);
      this.notify('Оценка сохранена', 'success');
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async openTask(task) {
    this.selectedTask = task;
    this.rateDraft = task.quality_rating || 0;
    try { this.taskComments = await api.get(`/tasks/tasks/${task.id}/comments`) || []; }
    catch { this.taskComments = []; }
    try { this.taskAssignees = await api.get(`/projects/tasks/${task.id}/assignees`) || []; }
    catch { this.taskAssignees = []; }
  },

  async addTaskAssignee(uid) {
    if (!uid || !this.selectedTask) return;
    try {
      await api.post(`/projects/tasks/${this.selectedTask.id}/assignees`, { user_id: uid });
      this.taskAssignees = await api.get(`/projects/tasks/${this.selectedTask.id}/assignees`) || [];
    } catch (e) { this.notify(e.message, 'error'); }
  },
  async removeTaskAssignee(uid) {
    if (!this.selectedTask) return;
    try {
      await api.del(`/projects/tasks/${this.selectedTask.id}/assignees/${uid}`);
      this.taskAssignees = await api.get(`/projects/tasks/${this.selectedTask.id}/assignees`) || [];
    } catch (e) { this.notify(e.message, 'error'); }
  },
  availableAssignees() {
    const used = new Set([...(this.taskAssignees || [])]);
    if (this.selectedTask?.assigned_to) used.add(this.selectedTask.assigned_to);
    return (this.participants || []).filter(p => !used.has(p.user_id));
  },

  canEditAttachment() {
    if (!this.selectedTask || !this.user) return false;
    if (this.canManage) return true;
    if (this.selectedTask.assigned_to === this.user.id) return true;
    return (this.taskAssignees || []).includes(this.user.id);
  },

  async uploadTaskAttachment(ev) {
    const file = ev.target.files && ev.target.files[0];
    if (!file || !this.selectedTask) { try { ev.target.value = ''; } catch {}; return; }
    try {
      const res = await api.upload(file);
      if (!res || !res.url) throw new Error('Сервер не вернул URL файла');
      await api.put(`/tasks/tasks/${this.selectedTask.id}/attachment`, { attachment_url: res.url });
      this.selectedTask.attachment_url = res.url;
      this.notify('Файл прикреплён: ' + (file.name || ''), 'success');
      await this.loadTasks();
    } catch (e) {
      this.notify('Ошибка загрузки: ' + (e.message || 'Unknown'), 'error');
    } finally {
      try { ev.target.value = ''; } catch {}
    }
  },

  async saveTaskAttachmentUrl() {
    if (!this.selectedTask) return;
    const url = (this.selectedTask.attachment_url || '').trim();
    if (url && !window.isValidUrl(url)) {
      this.notify('Введите валидную ссылку https:// или загрузите файл', 'error');
      return;
    }
    try {
      await api.put(`/tasks/tasks/${this.selectedTask.id}/attachment`, { attachment_url: url });
      this.notify(url ? 'Ссылка сохранена' : 'Вложение удалено', 'success');
      await this.loadTasks();
    } catch (e) { this.notify(e.message, 'error'); }
  },

  async clearTaskAttachment() {
    if (!this.selectedTask) return;
    try {
      await api.put(`/tasks/tasks/${this.selectedTask.id}/attachment`, { attachment_url: '' });
      this.selectedTask.attachment_url = '';
      this.notify('Вложение удалено', 'success');
      await this.loadTasks();
    } catch (e) { this.notify(e.message, 'error'); }
  },

  closeTask() { this.selectedTask = null; },

  async addComment() {
    if (!this.newComment.trim() || !this.selectedTask) return;
    try {
      await api.post(`/tasks/tasks/${this.selectedTask.id}/comments`, { content: this.newComment });
      this.taskComments = await api.get(`/tasks/tasks/${this.selectedTask.id}/comments`) || [];
      this.newComment = '';
    } catch (e) { this.notify(e.message, 'error'); }
  },

  dragStart(ev, task) { ev.dataTransfer.setData('text/plain', task.id); ev.dataTransfer.effectAllowed = 'move'; },
  dragOver(ev) { ev.preventDefault(); ev.dataTransfer.dropEffect = 'move'; },
  drop(ev, status) {
    ev.preventDefault();
    const id = ev.dataTransfer.getData('text/plain');
    const task = this.tasks.find(t => t.id === id);
    if (!task || task.status === status) return;
    const isMyTask = task.assigned_to === this.user?.id || (this.taskAssignees || []).includes(this.user?.id);
    if (this.canManage) {
    } else if (isMyTask) {
      if (status === 'новая' || status === 'завершена') {
        this.notify('Завершить задачу может только заместитель или руководитель', 'error');
        return;
      }
      if (task.status === 'на проверке' && status === 'в работе') {
        this.notify('Только заместитель или руководитель возвращает задачу из «на проверке»', 'error');
        return;
      }
    } else {
      this.notify('Менять статус может только исполнитель, заместитель или руководитель', 'error');
      return;
    }
    this.setTaskStatus(task, status);
  },

  participantName(userId) {
    const p = this.participants.find(x => x.user_id === userId);
    if (!p) return 'не назначен';
    return ((p.first_name || '') + ' ' + (p.last_name || '')).trim() || p.email || '';
  },

  priorityIcon(p) { return p === 'высокий' ? '!!!' : p === 'средний' ? '!!' : '!'; },
  priorityClasses(p) {
    if (p === 'высокий') return 'bg-gradient-to-br from-rose-50 via-white to-white border-rose-200';
    if (p === 'средний') return 'bg-gradient-to-br from-amber-50 via-white to-white border-amber-200';
    return 'bg-white border-ink-200';
  },
  priorityBadge(p) {
    if (p === 'высокий') return 'text-rose-600 bg-rose-100';
    if (p === 'средний') return 'text-amber-700 bg-amber-100';
    return 'text-emerald-700 bg-emerald-100';
  },
};
