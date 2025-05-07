var currentSessionId = null;
var currentActiveTabId = null;

// Function to handle deleting a session
async function handleDeleteSession(event) {
    event.stopPropagation(); // Prevent triggering session load if clicking on the icon within the item
    const sessionItem = event.currentTarget.closest('.session-item'); // Find the parent session item
    const sessionId = event.currentTarget.getAttribute('data-id');

    // Use the custom confirm dialog
    const confirmed = await confirmDialog('Are you sure you want to delete this chat session? This action cannot be undone.');

    if (confirmed) {
        activeSession = await apiDeleteSession(sessionId)

        if(activeSession){
            if(activeSession.id != currentSessionId){
                console.log("Deleted the active session. Resetting chat view and connection.");
                currentSessionId = activeSession.id;
                chatChangeSession(activeSession);
                updateSessionHighlight(currentSessionId);
            }
            if (sessionItem) {
                sessionItem.remove();
            }
        } else {
            appendMessage("Failed to deleted session", "error")
        }
    }
}

// Function to create a new chat session
async function createNewSession() {
    const session = await apiCreateSession();

    if (session) {
        console.log("New session created and connected:", session.id);

        currentSessionId = session.id;
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

        updateSessionHighlight(currentSessionId);

    } else {
        appendMessage("Error: Could not create new session.", "error");
    }
}

async function populateSessions() {
    sessions = await apiListSessions()

    if(sessions){
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

        updateSessionHighlight(currentSessionId);
    }
}

function updateSessionHighlight(activeSessionId) {
    const sessionItems = document.querySelectorAll('.session-item');
    sessionItems.forEach(item => {
        if (item.getAttribute('data-id') === activeSessionId) {
            item.classList.add('active');
        } else {
            item.classList.remove('active');
        }
    });
}

// Add event listeners for the new UI elements
document.getElementById('reloadSessions').addEventListener('click', populateSessions);
document.getElementById('newSession').addEventListener('click', createNewSession);

async function handleSessionClick(event) {
    const sessionItem = event.currentTarget;
    const sessionId = sessionItem.getAttribute('data-id');

    try {
        const response = await fetch(`http://localhost:8008/agent/sessions/connect/${sessionId}`, {
            method: 'GET',
            headers: { 'Accept': 'application/json' },
        });

        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }

        const data = await response.json();

        if (data && data.session && data.session.id) {
            currentSessionId = data.session.id;
            console.log("Connected to session:", currentSessionId);
            updateSessionHighlight(currentSessionId)

            // Clear chat view
            chatView.innerHTML = '';

            // Replace messages in chat view with the new session's messages
            if (data.session.messages && Array.isArray(data.session.messages)) {
                data.session.messages.forEach(message => {
                    let messageType = message.origin;
                    if (message.text) {
                        appendMessage(message.text, messageType);
                    }
                });
                scrollToBottom();
            } else {
                appendMessage("Connected, but no previous messages found.", "system");
            }
        } else {
            appendMessage("Error: Could not establish a session ID.", "error");
            currentSessionId = null;
        }
    } catch (error) {
        console.error("Failed to connect to session:", error);
        appendMessage(`Error connecting to session: ${error.message}`, "error");
        currentSessionId = null;
    }
}