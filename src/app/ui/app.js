const tabContents = document.querySelectorAll('.tab-content');
const sendButton = document.getElementById('sendButton');
const contentArea = document.querySelector('.content-area');
const modelSelector = document.getElementById('modelSelector');
const topTabButtons = document.querySelectorAll('.top-tab-button');

var currentSessionId = null;
var currentActiveTabId = null;

// --- Function to handle tab switching and panel toggle ---
function handleTopTabClick(event) {
    const clickedButton = event.currentTarget;
    const targetTabId = clickedButton.getAttribute('data-tab');
    const isPanelOpen = contentArea.classList.contains('side-panel-open');
    const isCurrentActiveButton = clickedButton.classList.contains('active');

    // --- Logic: Click active button when panel is open -> Close panel ---
    if (isPanelOpen && isCurrentActiveButton) {
        contentArea.classList.remove('side-panel-open');
        clickedButton.classList.remove('active');
        currentActiveTabId = null;
    }
    // --- Logic: Click any button when panel is closed, or different button when open -> Open/Switch tab ---
    else {
        // 1. Open panel if closed
        if (!isPanelOpen) {
            contentArea.classList.add('side-panel-open');
        }

        // 2. Update button active state
        topTabButtons.forEach(btn => btn.classList.remove('active'));
        clickedButton.classList.add('active');

        // 3. Update content active state
        tabContents.forEach(content => {
            if (content.id === targetTabId) {
                content.classList.add('active');
            } else {
                content.classList.remove('active');
            }
        });
        currentActiveTabId = targetTabId;
    }
}

// --- Initial Connection on Load ---
document.addEventListener('DOMContentLoaded', async () => {
    console.log("App loaded. Attempting to connect to agent...");

    populateModelSelector();
    populateSessions();

    session = await apiConnectSession()
    if(session){
        currentSessionId = session.id
        chatChangeSession(session)
    }

    // --- Set initial active tab state based on HTML ---
    const initiallyActiveButton = document.querySelector('.top-tab-button.active');
    if (initiallyActiveButton) {
        currentActiveTabId = initiallyActiveButton.getAttribute('data-tab');
        // Ensure corresponding content is visible
        tabContents.forEach(content => {
            content.classList.toggle('active', content.id === currentActiveTabId);
        });
        if (currentActiveTabId) {
            contentArea.classList.add('side-panel-open');
        }
    } else if (topTabButtons.length > 0) {
        const activeContent = document.querySelector('.tab-content.active');
        if (activeContent) {
            currentActiveTabId = activeContent.id;
            // Find the corresponding button and mark it active
            const correspondingButton = document.querySelector(`.top-tab-button[data-tab="${currentActiveTabId}"`);
            if (correspondingButton) {
                correspondingButton.classList.add('active');
                contentArea.classList.add('side-panel-open'); // Open panel if content is active
            }
        } else {
            // If no button and no content is active, ensure panel is closed.
            contentArea.classList.remove('side-panel-open');
        }
    }
});

// --- Add New Top Tab Button Listener ---
topTabButtons.forEach(button => {
    button.addEventListener('click', handleTopTabClick);
});

// --- Send Button Logic ---
sendButton.addEventListener('click', sendMessage);

async function populateModelSelector() {
    data = await apiListModels()
    if(data){
        modelSelector.innerHTML = '<option value="" disabled>Select a Model</option>'; // Keep placeholder

        let activeModelFound = false;
        data.models.forEach(model => {
            const option = document.createElement('option');
            option.value = model.id;
            option.textContent = model.name;
            if (model.id === data.activeModelID) {
                option.selected = true;
                activeModelFound = true;
            }
            modelSelector.appendChild(option);
        });

        // If the active model wasn't in the list or no active model was specified, select the placeholder
        if (!activeModelFound) {
            const placeholder = modelSelector.querySelector('option[disabled]');
            if (placeholder) {
                placeholder.selected = true;
            }
        }
    } else {
        modelSelector.innerHTML = '<option value="" disabled selected>Error loading models</option>';
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

        // Add event listeners for delete icons
        document.querySelectorAll('.delete-icon').forEach(icon => {
            icon.addEventListener('click', handleDeleteSession);
        });
    }
}

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
    try {
        const response = await fetch('http://localhost:8008/agent/sessions/new');
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();

        if (data && data.session && data.session.id && data.session.summary && data.session.date) {
            console.log("New session created:", data);

            // Create a new session item
            const sessionItem = document.createElement('div');
            sessionItem.classList.add('session-item');
            sessionItem.setAttribute("data-id", data.session.id);

            const summary = data.summary ? data.summary : 'New chat';
            sessionItem.innerHTML = `
                <span class="session-summary">${summary}</span>
                <img src="icons/delete.svg" alt="Delete" class="delete-icon" data-id="${data.id}">
            `;

            // Insert the new session item at the top of the session list
            const sessionList = document.getElementById('sessionList');
            sessionList.insertBefore(sessionItem, sessionList.firstChild);

            // Add event listener for the delete icon
            sessionItem.querySelector('.delete-icon').addEventListener('click', handleDeleteSession);
        } else {
            console.error("Invalid data structure received for new session:", data);
        }
    } catch (error) {
        console.error("Failed to create new session:", error);
    }
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

            // Clear chat view
            chatView.innerHTML = '';

            // Replace messages in chat view with the new session's messages
            if (data.session.messages && Array.isArray(data.session.messages)) {
                data.session.messages.forEach(message => {
                    let messageType = message.origin;
                    if (message.text) {
                        appendMessage(message.text, messageType);
                    } else {
                        console.warn("Message missing text:", message);
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