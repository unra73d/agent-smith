'use strict'

class SessionList extends List {
    constructor() {
        super();

        document.addEventListener('sessions:new', async e => await this.createNewSession());
        document.addEventListener('sessions:touch', e => this.touch());

        document.addEventListener('storage:sessions', e => this.items = Storage.sessions || []);
        document.addEventListener('session:update', e => this.items = Storage.sessions || []);
        document.addEventListener('storage:current-session', e => this.updateSessionHighlight());

        this._initStyle();
    }

    async _initStyle() {
        await super._initStyle();
        this.shadowRoot.adoptedStyleSheets = [
            ...this.shadowRoot.adoptedStyleSheets,
            await loadCSS('components/sessions/sessions.css')
        ];
    }

    getItem(session) {
        const item = document.createElement('div');
        item.classList.add('session-item');
        item.setAttribute("data-id", session.id);

        if (Storage.currentSession && session.id == Storage.currentSession.id) {
            item.classList.add('active');
        }

        const summary = session.summary ? session.summary : 'New chat';
        item.innerHTML = `
            <span class="session-summary">${summary}</span>
            <div alt="Delete" class="delete-icon img-button" data-id="${session.id}">&#xe053;</div>
        `;

        item.querySelector('.delete-icon').addEventListener('click', e => this.handleDeleteSession(e, session.id));
        item.addEventListener('click', e => this.onItemClick(item, session))
        return item;
    }

    onItemClick(item, session) {
        Storage.currentSession = session;
    }

    async handleDeleteSession(e, sessionId) {
        e.stopPropagation();
        const currSessionId = Storage.currentSession.id;

        const confirmed = await confirmDialog('Delete this chat session?');
        if (confirmed) {
            await apiDeleteSession(sessionId);

            Storage.sessions = Storage.sessions.filter(s => s.id !== sessionId);

            if (currSessionId == sessionId) {
                if (Storage.sessions.length == 0) {
                    await this.createNewSession();
                } else {
                    Storage.currentSession = Storage.sessions[0];
                }
            }
        }
    }

    async createNewSession() {
        const session = await apiCreateSession();
        if (session) {
            Storage.sessions = [session, ...(Storage.sessions || [])];
            Storage.currentSession = session;
        }
    }

    updateSessionHighlight() {
        for (let item of this.list.children) {
            if (item.getAttribute('data-id') === Storage.currentSession.id) {
                item.classList.add('active');
            } else {
                item.classList.remove('active');
            }
        }
    }

    touch() {
        if (Storage.currentSession) {
            const currentIndex = Storage.sessions.findIndex(session => session.id === Storage.currentSession.id);
            if (currentIndex > 0) {
                Storage.sessions = [
                    Storage.currentSession,
                    ...Storage.sessions.slice(0, currentIndex),
                    ...Storage.sessions.slice(currentIndex + 1)
                ];
                Storage.currentSession.date = new Date().toISOString();
                this.items = Storage.sessions;
                this.updateSessionHighlight();
            }
        }
    }
}

customElements.define('session-list', SessionList);