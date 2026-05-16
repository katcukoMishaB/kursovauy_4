function app() {
  const modules = {
    AppState: window.AppState ? window.AppState() : null,
    AuthModule: window.AuthModule,
    ProjectsModule: window.ProjectsModule,
    TasksModule: window.TasksModule,
    ChatModule: window.ChatModule,
    ProfileModule: window.ProfileModule,
    AdminModule: window.AdminModule,
    RouterModule: window.RouterModule,
    InvitationsModule: window.InvitationsModule,
  };
  for (const [k, v] of Object.entries(modules)) {
    if (!v) console.error('[app] module missing:', k);
  }
  return Object.assign({}, ...Object.values(modules).filter(Boolean));
}
window.app = app;
