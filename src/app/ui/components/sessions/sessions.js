'use strict'

class SessionList extends HTMLElement {
    constructor() {
        super()
        const shadowRoot = this.attachShadow({ mode: 'open' })

        const styles = document.createElement('style');
        styles.innerHTML = `
        @import url('global.css');
        @import url('components/sessions/sessions.css');
        `
        shadowRoot.appendChild(styles);

        this.list = document.createElement('div')
        shadowRoot.appendChild(this.list)

        this.createNewSession = this.createNewSession.bind(this)

        document.addEventListener('sessions:new', async e => await this.createNewSession())
        document.addEventListener('sessions:touch', e => this.touch())

        document.addEventListener('storage:sessions', e => this.updateList())
        document.addEventListener('storage:current-session', e => this.updateSessionHighlight())
    }

    updateList() {
        console.debug("session list update")
        if (Storage.sessions && Storage.sessions.length > 0) {
            this.list.innerHTML = ''

            for (let session of Storage.sessions) {
                this.appendSession(session)
            }
        }
    }

    appendSession(session, front = false) {
        const sessionItem = document.createElement('div');
        sessionItem.classList.add('session-item');
        sessionItem.setAttribute("data-id", session.id);
        if (Storage.currentSession && session.id == Storage.currentSession.id) {
            sessionItem.classList.add('active')
        }

        const summary = session.summary ? session.summary : 'New chat';
        sessionItem.innerHTML = `
            <span class="session-summary">${summary}</span>
            <div alt="Delete" class="delete-icon img-button" data-id="${session.id}">🗑️</div>
        `;

        if (front) {
            this.list.insertBefore(sessionItem, this.list.firstChild);
        } else {
            this.list.appendChild(sessionItem);
        }

        sessionItem.addEventListener('click', e => Storage.currentSession = session)
        sessionItem.querySelector('.delete-icon').addEventListener('click', e => this.handleDeleteSession(e, session.id))
    }

    async handleDeleteSession(e, sessionId) {
        e.stopPropagation()
        const currSessionId = Storage.currentSession.id

        // Use the custom confirm dialog
        const confirmed = await confirmDialog('Are you sure you want to delete this chat session? This action cannot be undone.');

        if (confirmed) {
            await apiDeleteSession(sessionId)

            for (let i in Storage.sessions) {
                if (Storage.sessions[i].id == sessionId) {
                    Storage.sessions = [...Storage.sessions.slice(0, i), ...Storage.sessions.slice(i + 1)]
                    break
                }
            }

            if (currSessionId == sessionId) {
                console.log("Deleted the active session. Resetting chat view and connection.");
                if (Storage.sessions.length == 0) {
                    this.createNewSession();
                } else {
                    Storage.currentSession = Storage.sessions[0];
                }

            }
        }
    }

    async createNewSession() {
        const session = await apiCreateSession();

        if (session) {
            console.log("New session created:", session.id);

            Storage.sessions = [session, ...Storage.sessions]
            Storage.currentSession = session

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
            // Find the index of the current session in the sessions array
            const currentIndex = Storage.sessions.findIndex(session => session.id === Storage.currentSession.id);

            // If the current session is not the first in the list
            if (currentIndex > 0) {
                // Move the current session to the beginning of the sessions array
                Storage.sessions = [
                    Storage.currentSession,
                    ...Storage.sessions.slice(0, currentIndex),
                    ...Storage.sessions.slice(currentIndex + 1)
                ];

                // Update the current session's date to now
                Storage.currentSession.date = new Date().toISOString();

                // Update the session list in the UI
                for (let i in this.list.children) {
                    let item = this.list.children[i]
                    if (item.getAttribute('data-id') === Storage.currentSession.id) {
                        this.list.removeChild(item)
                        this.list.insertBefore(item, this.list.firstChild);
                        break
                    }
                }

                // Update the highlight to reflect the new order
                this.updateSessionHighlight();
            }
        }
    }
}

customElements.define('session-list', SessionList)

