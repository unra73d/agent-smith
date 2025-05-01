const sandwichButton = document.getElementById('sandwichButton');
const sidePanel = document.getElementById('sidePanel');
const mainChat = document.getElementById('mainChat');
const appContainer = document.querySelector('.app-container');
const chatInput = document.getElementById('chatInput');
const tabButtons = document.querySelectorAll('.tab-button');
const tabContents = document.querySelectorAll('.tab-content');
const sendButton = document.getElementById('sendButton');
const chatView = document.getElementById('chatView');
const contentArea = document.querySelector('.content-area');


let currentSessionId = null; // Variable to store the session ID

// --- Function to append messages to the chat view ---
function appendMessage(text, type) {
    const messageElement = document.createElement('div');
    messageElement.classList.add('message', type);

    // NOTE: Basic text handling for now.
    // In the future, this is where you'd parse markdown,
    // identify code blocks, tables, images etc., and render them correctly.
    // For now, just setting textContent for simplicity.
    const textNode = document.createTextNode(text);
    messageElement.appendChild(textNode);

    // TODO: Add handling for code blocks, tables, images based on backend response format

    chatView.appendChild(messageElement);

    // Scroll to the bottom of the chat view
    scrollToBottom();
}

// --- Function to scroll chat view to the bottom ---
function scrollToBottom() {
    chatView.scrollTop = chatView.scrollHeight;
}


// --- Initial Connection on Load ---
document.addEventListener('DOMContentLoaded', async () => {
    console.log("App loaded. Attempting to connect to agent...");
    try {
        const response = await fetch('http://localhost:8008/agent/connect', {
            method: 'GET', 
            headers: { 'Accept': 'application/json' }, // Expect JSON response
        });

        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }

        const data = await response.json();

        if (data && data.session && data.session.id) {
            currentSessionId = data.session.id;
            console.log("Connected to session:", currentSessionId);

            if (data.session.messages && Array.isArray(data.session.messages)) {
                // Clear any initial "connecting" message if needed
                // chatView.innerHTML = ''; // Optional: Clear view before adding history

                data.session.messages.forEach(message => {
                    let messageType = 'system'; // Default type

                    // Map backend origin constants to CSS classes
                    // const (
                    //     MessageOriginUser = 1
                    //     MessageOriginAI   = 2
                    //     MessageOriginTool = 3
                    // )
                    switch (message.origin) {
                        case 1:
                            messageType = 'user';
                            break;
                        case 2:
                            messageType = 'assistant';
                            break;
                        case 3:
                            messageType = 'tool';
                            break;
                        default:
                            console.warn("Unknown message origin:", message.origin);
                            messageType = 'system'; // Fallback for unknown origins
                    }

                    // Assuming message structure is { content: "...", origin: X }
                    if (message.text) {
                        appendMessage(message.text, messageType);
                    } else {
                        console.warn("Message missing text:", message);
                    }
                });
                scrollToBottom(); // Scroll after loading history
            } else {
                appendMessage("Connected, but no previous messages found.", "system");
            }
        } else {
            appendMessage("Error: Could not establish a session ID.", "error");
            currentSessionId = null; // Ensure it's null if connection failed
        }

    } catch (error) {
        console.error('Error connecting to agent:', error);
        appendMessage(`Error connecting to agent: ${error.message}. Please ensure the backend is running.`, "error");
        // Disable input if connection fails?
        // chatInput.disabled = true;
        // sendButton.disabled = true;
    }
});


// --- Side Panel Toggle ---
sandwichButton.addEventListener('click', () => {
    contentArea.classList.toggle('side-panel-open');
});

// --- Textarea Auto-Resize ---
const initialHeight = chatInput.scrollHeight; // Store initial height for reset if needed

chatInput.addEventListener('input', () => {
    chatInput.style.height = 'auto'; // Reset height to recalculate
    const scrollHeight = chatInput.scrollHeight;
    const maxHeight = 150; // Match max-height in CSS

    // Set height based on content, but not exceeding max-height
    chatInput.style.height = `${Math.min(scrollHeight, maxHeight)}px`;

    // Show scrollbar if content exceeds max-height
    chatInput.style.overflowY = scrollHeight > maxHeight ? 'auto' : 'hidden';
});

// Reset height if input is cleared (optional)
chatInput.addEventListener('blur', () => { // Or use a clear button
    if (chatInput.value === '') {
        chatInput.style.height = 'auto'; // Use default rows="1" height
        chatInput.style.overflowY = 'hidden';
    }
});


// --- Side Panel Tab Switching ---
tabButtons.forEach(button => {
    button.addEventListener('click', () => {
        const targetTabId = button.getAttribute('data-tab');

        // Update button active state
        tabButtons.forEach(btn => btn.classList.remove('active'));
        button.classList.add('active');

        // Update content active state
        tabContents.forEach(content => {
            if (content.id === targetTabId) {
                content.classList.add('active');
            } else {
                content.classList.remove('active');
            }
        });
    });
});


// --- Send Button Logic ---
sendButton.addEventListener('click', async () => {
    const messageText = chatInput.value.trim();
    if (!messageText) {
        return; // Don't send empty messages
    }
    if (!currentSessionId) {
        appendMessage("Error: Not connected to agent. Cannot send message.", "error");
        return;
    }

    // 1. Display user message in chatView
    appendMessage(messageText, 'user'); // Use the existing function

    // 2. Clear input and reset height
    chatInput.value = '';
    chatInput.style.height = 'auto';
    chatInput.style.overflowY = 'hidden';
    chatInput.focus(); // Keep focus on input

    // 3. Send messageText to backend API, including the session ID
    console.log("Sending message with session ID:", currentSessionId);

    // displayTypingIndicator(); // Show typing indicator (implement if needed)
    try {
        // Example: POST request to backend
        const response = await fetch('http://localhost:8008/agent/directchat', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                session_id: currentSessionId, // Include the session ID
                message: messageText
            }),
        });

        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);

        const data = await response.json();

        // 4. Display assistant response
        // Assuming the backend response structure is { "response": "..." }
        if (data && data.response) {
             appendMessage(data.response, 'assistant');
        } else {
            console.error("Received unexpected response format:", data);
            appendMessage("Error: Received an unexpected response from the server.", "error");
        }


    } catch (error) {
        console.error('Error calling chat API:', error);
        appendMessage(`Error sending message: ${error.message}`, 'error'); // Display error in chat
    } finally {
         // removeTypingIndicator(); // Hide typing indicator (implement if needed)
    }
});
