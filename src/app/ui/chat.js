const chatInput = document.getElementById('chatInput');
const chatView = document.getElementById('chatView');

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

chatInput.addEventListener('keydown', (event) => {
    // Check for Enter key without Shift modifier
    if (event.key === 'Enter' && !event.shiftKey) {
        event.preventDefault();
        sendMessage();
    }
});

function chatChangeSession(session){
    if (session.messages && Array.isArray(session.messages)) {
        chatView.innerHTML = '';
        session.messages.forEach(message => {
            appendMessage(message.text, message.origin);
        });
        scrollToBottom();
    }
}