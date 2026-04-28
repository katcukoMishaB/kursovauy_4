function appShell() {
    return {
        page: 'home',
        init() {
            const originalShowPage = window.Utils.showPage;
            window.Utils.showPage = (pageId) => {
                this.page = pageId;
                originalShowPage(pageId);
            };
            window.Accessibility.loadAccessibilitySettings();
            window.Accessibility.applyAccessibilitySettings();
            window.Auth.checkAuth();
        },
        go(pageId) {
            window.Utils.showPage(pageId);
        },
        login(event) {
            window.Auth.login(event);
        },
        register(event) {
            window.Auth.register(event);
        },
        logout() {
            window.Auth.logout();
        }
    };
}

window.appShell = appShell;
