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

        this.createNewSession = this.createNewSession.bind(this)

        document.addEventListener('sessions:new', e=>this.createNewSession())
        document.addEventListener('sessions:touch', e=>this.touch())

        document.addEventListener('storage:sessions', e=>this.updateList())
        document.addEventListener('storage:current-session', e=>this.updateSessionHighlight())
    }

    updateList(){
        if(Storage.sessions && Storage.sessions.length > 0){
            this.list.innerHTML = ''

            for (let session of Storage.sessions) {
                this.appendSession(session)
            }
        }
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

        sessionItem.addEventListener('click', e=>Storage.currentSession=session)
        sessionItem.querySelector('.delete-icon').addEventListener('click', e=>this.handleDeleteSession(e, sessionItem, session.id))
    }

    async handleDeleteSession(e, sessionItem, sessionId){
        e.stopPropagation()
        const currSessionId = Storage.currentSession.id

        // Use the custom confirm dialog
        const confirmed = await confirmDialog('Are you sure you want to delete this chat session? This action cannot be undone.');

        if (confirmed) {
            await apiDeleteSession(sessionId)
            
            this.list.removeChild(sessionItem)

            for(let i in Storage.sessions){
                if(Storage.sessions[i] == Storage.currentSession){
                    Storage.sessions.splice(i, 1)
                    break
                }
            }

            if(currSessionId == sessionId){
                console.log("Deleted the active session. Resetting chat view and connection.");
                if (Storage.sessions.length == 0){
                    this.createNewSession();
                } else {
                    Storage.currentSession = Storage.sessions[0];
                    this.updateSessionHighlight(Storage.currentSession);
                }
                
            }
        }
    }

    async createNewSession() {
        const session = await apiCreateSession();
    
        if (session) {
            console.log("New session created:", session.id);
    
            Storage.sessions.unshift(session)
            this.appendSession(session, true)
            this.selectSession(session.id)

            this.updateSessionHighlight(Storage.currentSession);

        } else {
            appendMessage("Error: Could not create new session.", "error");
        }
    }

    updateSessionHighlight() {
        for(let item of this.list.children){
            if (item.getAttribute('data-id') === Storage.currentSession.id) {
                item.classList.add('active');
            } else {
                item.classList.remove('active');
            }
        }
    }

    selectSession(sessionId){
        for (let session of Storage.sessions){
            if(session.id == sessionId){
                Storage.currentSession = session
                break
            }
        }
        this.updateSessionHighlight()
    }

    touch() {
        if (Storage.currentSession) {
            // Find the index of the current session in the sessions array
            const currentIndex = Storage.sessions.findIndex(session => session.id === Storage.currentSession.id);
    
            // If the current session is not the first in the list
            if (currentIndex > 0) {
                // Move the current session to the beginning of the sessions array
                Storage.sessions.splice(currentIndex, 1);
                Storage.sessions.unshift(Storage.currentSession);
    
                // Update the current session's date to now
                Storage.currentSession.date = new Date().toISOString();
    
                // Update the session list in the UI
                for(let i in this.list.children){
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

