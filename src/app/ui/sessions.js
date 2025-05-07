var currentSession = null;
var sessions = []

// Function to handle deleting a session
async function handleDeleteSession(event) {
    event.stopPropagation(); // Prevent triggering session load if clicking on the icon within the item
    const sessionItem = event.currentTarget.closest('.session-item'); // Find the parent session item
    const sessionId = event.currentTarget.getAttribute('data-id');
    const currentSessionId = currentSession.id

    // Use the custom confirm dialog
    const confirmed = await confirmDialog('Are you sure you want to delete this chat session? This action cannot be undone.');

    if (confirmed) {
        await apiDeleteSession(sessionId)
        sessionItem.remove();
        for(i in sessions){
            if(sessions[i] == currentSession){
                sessions.splice(i, 1)
                break
            }
        }

        if(currentSessionId == sessionId){
            console.log("Deleted the active session. Resetting chat view and connection.");
            if (sessions.length == 0){
                createNewSession();
            } else {
                currentSession = sessions[0];
                chatChangeSession(currentSession);
                updateSessionHighlight(currentSession);
            }
            
        }
    }
}

// Function to create a new chat session
async function createNewSession() {
    const session = await apiCreateSession();

    if (session) {
        console.log("New session created and connected:", session.id);

        sessions.unshift(session)
        currentSession = session;
        chatChangeSession(session);

        // Create and add the new session item to the UI list
        const sessionList = document.getElementById('sessionList');
        const sessionItem = document.createElement('div');

        sessionItem.classList.add('session-item');
        sessionItem.setAttribute("data-id", session.id);

        const summary = session.summary ? session.summary : 'New chat';
        sessionItem.innerHTML = `
            <span class="session-summary">${summary}</span>
            <img src="icons/delete.svg" alt="Delete" class="delete-icon" data-id="${session.id}">
        `;
        sessionList.insertBefore(sessionItem, sessionList.firstChild);

        sessionItem.addEventListener('click', handleSessionClick);
        sessionItem.querySelector('.delete-icon').addEventListener('click', handleDeleteSession);

        updateSessionHighlight(currentSession);

    } else {
        appendMessage("Error: Could not create new session.", "error");
    }
}

async function populateSessions() {
    sessions = await apiListSessions()

    if(sessions){
        sessions.sort((a, b) => new Date(b.date) - new Date(a.date));
        const sessionList = document.getElementById('sessionList');
        sessionList.innerHTML = '';

        sessions.forEach(session => {
            const sessionItem = document.createElement('div');
            sessionItem.classList.add('session-item');
            sessionItem.setAttribute("data-id", session.id);

            const summary = session.summary ? session.summary : 'New chat';
            sessionItem.innerHTML = `
                <span class="session-summary">${summary}</span>
                <img src="icons/delete.svg" alt="Delete" class="delete-icon" data-id="${session.id}">
            `;

            sessionList.appendChild(sessionItem);
        });

        document.querySelectorAll('.session-item').forEach(item => {
            item.addEventListener('click', handleSessionClick);
        });

        document.querySelectorAll('.delete-icon').forEach(icon => {
            icon.addEventListener('click', handleDeleteSession);
        });

        if(sessions.length > 0){
            currentSession = sessions[0]
        }

        selectSession(currentSession.id);
    } else {
        sessions = []
    }
}

function updateSessionHighlight() {
    const sessionItems = document.querySelectorAll('.session-item');
    sessionItems.forEach(item => {
        if (item.getAttribute('data-id') === currentSession.id) {
            item.classList.add('active');
        } else {
            item.classList.remove('active');
        }
    });
}

// Add event listeners for the new UI elements
document.getElementById('reloadSessions').addEventListener('click', populateSessions);
document.getElementById('newSession').addEventListener('click', createNewSession);

function selectSession(sessionId){
    for (i in sessions){
        if(sessions[i].id == sessionId){
            currentSession = sessions[i]
            break
        }
    }
    updateSessionHighlight()

    if(currentSession != null){
        chatChangeSession(currentSession)
    }
}

function handleSessionClick(event) {
    const sessionItem = event.currentTarget;
    const sessionId = sessionItem.getAttribute('data-id');
    selectSession(sessionId)
}

function touchSession() {
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
            const sessionList = document.getElementById('sessionList');
            const sessionItems = sessionList.querySelectorAll('.session-item');
            sessionItems.forEach(item => {
                if (item.getAttribute('data-id') === currentSession.id) {
                    sessionList.insertBefore(item, sessionList.firstChild);
                }
            });

            // Update the highlight to reflect the new order
            updateSessionHighlight();
        }
    }
}