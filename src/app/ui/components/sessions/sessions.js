'use strict'

class SessionList extends HTMLElement {
    constructor() {
        super()
        const shadowRoot = this.attachShadow({mode: 'open'})
        
        const styles = document.createElement('style');
        styles.innerHTML = `
        @import url('global.css');
        @import url('components/sessions/sessions.css');
        `
        shadowRoot.appendChild(styles);

        this.list = document.createElement('div')
        shadowRoot.appendChild(this.list)

        this.populateSessions = this.populateSessions.bind(this)
        this.createNewSession = this.createNewSession.bind(this)

        document.addEventListener('sessions:new', e=>this.createNewSession())
        document.addEventListener('sessions:touch', e=>this.touch())
        document.addEventListener('sessions:reload', e=>this.populateSessions())
    }

    connectedCallback(){
        this.populateSessions()
    }

    appendSession(session, front=false){
        const sessionItem = document.createElement('div');
        sessionItem.classList.add('session-item');
        sessionItem.setAttribute("data-id", session.id);

        const summary = session.summary ? session.summary : 'New chat';
        sessionItem.innerHTML = `
            <span class="session-summary">${summary}</span>
            <img src="icons/delete.svg" alt="Delete" class="delete-icon" data-id="${session.id}">
        `;

        if(front){
            this.list.insertBefore(sessionItem, this.list.firstChild);
        } else {
            this.list.appendChild(sessionItem);
        }

        sessionItem.addEventListener('click', e=>this.selectSession(session.id))
        sessionItem.querySelector('.delete-icon').addEventListener('click', e=>this.handleDeleteSession(e, sessionItem, session.id))
    }

    async populateSessions(){
        sessions = await apiListSessions()

        if(sessions){
            sessions.sort((a, b) => new Date(b.date) - new Date(a.date));
            this.list.innerHTML = ''

            for (let session of sessions) {
                this.appendSession(session)
            }

            if(sessions.length > 0){
                currentSession = sessions[0]
            }

            this.selectSession(currentSession.id);
        } else {
            sessions = []
        }
    }

    async handleDeleteSession(e, sessionItem, sessionId){
        e.stopPropagation()
        const currentSessionId = currentSession.id

        // Use the custom confirm dialog
        const confirmed = await confirmDialog('Are you sure you want to delete this chat session? This action cannot be undone.');

        if (confirmed) {
            await apiDeleteSession(sessionId)
            
            this.list.removeChild(sessionItem)

            for(let i in sessions){
                if(sessions[i] == currentSession){
                    sessions.splice(i, 1)
                    break
                }
            }

            if(currentSessionId == sessionId){
                console.log("Deleted the active session. Resetting chat view and connection.");
                if (sessions.length == 0){
                    this.createNewSession();
                } else {
                    currentSession = sessions[0];
                    sendEvent('chat:change-session', {session: currentSession})
                    this.updateSessionHighlight(currentSession);
                }
                
            }
        }
    }

    async createNewSession() {
        const session = await apiCreateSession();
    
        if (session) {
            console.log("New session created:", session.id);
    
            sessions.unshift(session)
            this.appendSession(session, true)
            this.selectSession(session.id)

            this.updateSessionHighlight(currentSession);

        } else {
            appendMessage("Error: Could not create new session.", "error");
        }
    }

    updateSessionHighlight() {
        for(let item of this.list.children){
            if (item.getAttribute('data-id') === currentSession.id) {
                item.classList.add('active');
            } else {
                item.classList.remove('active');
            }
        }
    }

    selectSession(sessionId){
        for (let session of sessions){
            if(session.id == sessionId){
                currentSession = session
                break
            }
        }
        this.updateSessionHighlight()

        if(currentSession != null){
            sendEvent('chat:change-session', {session: currentSession})
        }
    }

    touch() {
        if (currentSession) {
            // Find the index of the current session in the sessions array
            const currentIndex = sessions.findIndex(session => session.id === currentSession.id);
    
            // If the current session is not the first in the list
            if (currentIndex > 0) {
                // Move the current session to the beginning of the sessions array
                sessions.splice(currentIndex, 1);
                sessions.unshift(currentSession);
    
                // Update the current session's date to now
                currentSession.date = new Date().toISOString();
    
                // Update the session list in the UI
                for(let i in this.list.children){
                    let item = this.list.children[i]
                    if (item.getAttribute('data-id') === currentSession.id) {
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

