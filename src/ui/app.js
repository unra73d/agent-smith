const sidePanel = document.getElementById('sidePanel');
const mainChat = document.getElementById('mainChat');
const appContainer = document.querySelector('.app-container');
const chatInput = document.getElementById('chatInput');
const tabContents = document.querySelectorAll('.tab-content');
const sendButton = document.getElementById('sendButton');
const chatView = document.getElementById('chatView');
const contentArea = document.querySelector('.content-area');
const modelSelector = document.getElementById('modelSelector');
const topTabButtons = document.querySelectorAll('.top-tab-button');

let currentSessionId = null;
let currentActiveTabId = null;

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



function applySyntaxHighlighting(element) {
    if (typeof hljs !== 'undefined') {
        element.querySelectorAll('pre code').forEach((block) => {
            try {
                hljs.highlightElement(block);
            } catch (e) {
                console.error("Highlight.js error:", e, "on block:", block);
            }
        });
    } else {
        console.warn('Highlight.js not loaded. Skipping syntax highlighting.');
    }
}

function appendMessage(content, type) {
    const messageElement = document.createElement('div');
    messageElement.classList.add('message', type);

    if (type === 'assistant') {
        // --- Start Modification for Assistant Messages ---
        try {
            // 1. Replace <think>...</think> with structured divs
            let processedContent = content.replace(
                /<think>([\s\S]*?)<\/think>/g,
                (match, thinkingContent) => {
                    // Trim whitespace from thinking content for cleaner display
                    const trimmedContent = thinkingContent.trim();
                    // Only create block if content exists
                    if (trimmedContent) {
                        return `<div class="thinking-block">
                                    <div class="thinking-summary">🤔 Thinking...</div>
                                    <div class="thinking-content">${trimmedContent}</div>
                               </div>`;
                    }
                    return '';
                }
            );

            // Remove any leftover empty thinking tags (if any edge cases occurred)
            processedContent = processedContent.replace(/<think>\s*<\/think>/g, '');

            // 2. Parse the processed content as Markdown
            const htmlContent = marked.parse(processedContent, {
                gfm: true,
                breaks: true,
                mangle: false,
                headerIds: false
            });
            messageElement.innerHTML = htmlContent;

            // 3. Add interactivity to thinking blocks
            const thinkingBlocks = messageElement.querySelectorAll('.thinking-block');
            thinkingBlocks.forEach(block => {
                const summary = block.querySelector('.thinking-summary');
                if (summary) {
                    summary.addEventListener('click', () => {
                        block.classList.toggle('open');
                    });
                } else {
                    console.warn("Thinking block found without a summary element:", block);
                }
            });

            applySyntaxHighlighting(messageElement);

        } catch (e) {
            messageElement.textContent = content;
            const errorDiv = document.createElement('div');
            errorDiv.classList.add('error');
            errorDiv.style.fontSize = '0.8em';
            errorDiv.style.marginTop = '5px';
            errorDiv.textContent = "[UI Error: Failed to render Markdown]";
            messageElement.appendChild(errorDiv);
        }
    } else {
        messageElement.textContent = content;
    }

    chatView.appendChild(messageElement);
    scrollToBottom();
}

function scrollToBottom() {
    chatView.scrollTop = chatView.scrollHeight;
}

// --- Send Message Logic (Refactored) ---
async function sendMessage() {
    const messageText = chatInput.value.trim();
    if (!messageText) {
        return;
    }
    if (!currentSessionId) {
        appendMessage("Error: Not connected to agent. Cannot send message.", "error");
        return;
    }

    // 1. Display user message in chatView
    appendMessage(messageText, 'user');

    // 2. Clear input and reset height
    chatInput.value = '';
    chatInput.style.height = 'auto';
    chatInput.style.overflowY = 'hidden';
    chatInput.focus();

    // 3. Send messageText to backend API
    try {
        const response = await fetch('http://localhost:8008/agent/directchat', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                sessionID: currentSessionId,
                message: messageText
            }),
        });

        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);

        const data = await response.json();

        // 4. Display assistant response
        if (data && data.response) {
            appendMessage(data.response, 'assistant');
        } else {
            console.error("Received unexpected response format:", data);
            appendMessage("Error: Received an unexpected response from the server.", "error");
        }
    } catch (error) {
        console.error("Failed to send message:", error);
        // Optionally display error to user in the chat
        appendMessage(`Error sending message: ${error.message}`, "error");
        // Consider if you want to restore the user's input on failure
        chatInput.value = messageText;
    }
}

// --- Initial Connection on Load ---
document.addEventListener('DOMContentLoaded', async () => {
    console.log("App loaded. Attempting to connect to agent...");

    populateModelSelector();
    populateSessions();

    // --- Connect to agent (existing code) ---
    try {
        const response = await fetch('http://localhost:8008/agent/sessions/connect', {
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
        console.error("Failed to connect:", error);
        appendMessage(`Error connecting to agent: ${error.message}`, "error");
        currentSessionId = null;
    }

    // --- Set initial active tab state based on HTML ---
    const initiallyActiveButton = document.querySelector('.top-tab-button.active');
    if (initiallyActiveButton) {
        currentActiveTabId = initiallyActiveButton.getAttribute('data-tab');
        // Ensure corresponding content is visible
        tabContents.forEach(content => {
            content.classList.toggle('active', content.id === currentActiveTabId);
        });
        // Ensure the side panel reflects the initial active state if a tab is active
        if (currentActiveTabId) {
            contentArea.classList.add('side-panel-open');
        }
    } else if (topTabButtons.length > 0) {
        // Fallback: If no button has .active, keep panel closed initially
        // and don't explicitly activate the first tab unless clicked.
        // The first tab's content might still be visible if it has .active in HTML.
        // Check if any content is active by default
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


// --- Remove old sandwichButton listener ---
// sandwichButton.addEventListener('click', () => { ... });

// --- Textarea Auto-Resize ---
const initialHeight = chatInput.scrollHeight;

chatInput.addEventListener('input', () => {
    chatInput.style.height = 'auto';
    const scrollHeight = chatInput.scrollHeight;
    const maxHeight = 150;

    chatInput.style.height = `${Math.min(scrollHeight, maxHeight)}px`;

    chatInput.style.overflowY = scrollHeight > maxHeight ? 'auto' : 'hidden';
});

// Reset height if input is cleared
chatInput.addEventListener('blur', () => {
    if (chatInput.value === '') {
        chatInput.style.height = 'auto';
        chatInput.style.overflowY = 'hidden';
    }
});

// --- Add New Top Tab Button Listener ---
topTabButtons.forEach(button => {
    button.addEventListener('click', handleTopTabClick);
});

document.addEventListener('keydown', (event) => {
    const isModifier = event.metaKey || event.ctrlKey;

    if (!isModifier) {
        return;
    }

    let command = null;
    switch (event.key.toLowerCase()) {
        case 'c':
            command = 'copy';
            break;
        case 'v':
            command = 'paste';
            break;
        case 'x':
            command = 'cut';
            break;
        case 'a':
            command = 'selectAll';
            break;
        case 'z':
            if (event.shiftKey) {
                command = 'redo';
            } else {
                command = 'undo';
            }
            break
        case 'y':
            command = 'redo';
            break;
    }

    if (command) {
        try {
            if (document.execCommand(command)) {
                event.preventDefault();
            } else {
                console.warn(`document.execCommand('${command}') failed or did not apply.`);
            }
        } catch (e) {
            console.error(`Error executing document.execCommand('${command}'):`, e);
        }
    }
});

// --- Chat Input Key Handling ---
chatInput.addEventListener('keydown', (event) => {
    // Check for Enter key without Shift modifier
    if (event.key === 'Enter' && !event.shiftKey) {
        event.preventDefault();
        sendMessage();
    }
});


// --- Send Button Logic ---
sendButton.addEventListener('click', sendMessage);

async function populateModelSelector() {
    try {
        const response = await fetch('http://localhost:8008/agent/models/list');
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();

        if (data && data.models && data.activeModelID !== undefined) {
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
            console.error("Invalid data structure received for models:", data);
            modelSelector.innerHTML = '<option value="" disabled selected>Error loading models</option>';
        }
    } catch (error) {
        console.error("Failed to populate model selector:", error);
        modelSelector.innerHTML = '<option value="" disabled selected>Error loading models</option>';
    }
}

async function populateSessions() {
    try {
        const response = await fetch('http://localhost:8008/agent/sessions/list');
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();

        if (data && data.sessions && Array.isArray(data.sessions)) {
            const sessionList = document.getElementById('sessionList');
            sessionList.innerHTML = ''; // Clear existing sessions

            data.sessions.forEach(session => {
                const sessionItem = document.createElement('div');
                sessionItem.classList.add('session-item');

                const summary = session.summary ? session.summary : 'New chat';
                sessionItem.innerHTML = `
                    <span class="session-summary">${summary}</span>
                    <img src="icons/delete.svg" alt="Delete" class="delete-icon" data-id="${session.id}">
                `;

                sessionList.appendChild(sessionItem);
            });

            // Add event listeners for delete icons
            document.querySelectorAll('.delete-icon').forEach(icon => {
                icon.addEventListener('click', handleDeleteSession);
            });
        } else {
            console.error("Invalid data structure received for sessions:", data);
        }
    } catch (error) {
        console.error("Failed to fetch sessions:", error);
    }
}

// Function to handle deleting a session
async function handleDeleteSession(event) {
    const sessionId = event.currentTarget.getAttribute('data-id');
    if (confirm('Are you sure you want to delete this session?')) {
        try {
            const response = await fetch(`http://localhost:8008/agent/sessions/delete/${sessionId}`);
            if (!response.ok) {
                throw new Error(`HTTP error! Status: ${response.status}`);
            }
            populateSessions(); // Refresh the session list
        } catch (error) {
            console.error("Failed to delete session:", error);
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
