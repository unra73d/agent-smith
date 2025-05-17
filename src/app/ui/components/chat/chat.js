'use strict'

class ChatView extends HTMLElement {
    constructor() {
        super()

        this.chatSession = null
        this.toolsSelected = true

        const shadowRoot = this.attachShadow({ mode: 'open' })

        this.shadowRoot.innerHTML = `
        <style>
            @import url('global.css');
            @import url('components/chat/chat.css');
            @import url('components/chat/syntax-theme.min.css');
        </style>
        <div class="chat-view">
        </div>
        <div class="chat-input-area">
            <button class="cancel-button img-button">🚫 Stop</button >
            <div class="chat-input-container">
                <textarea id="chatInput" class="chat-input" placeholder="Enter your message..." rows="1"></textarea>
            </div>
            <div class="chat-button-container">
                <ui-checkbox class="tools-checkbox" label="Tools" ${this.toolsSelected ? 'checked' : ''}></ui-checkbox>
                <button id="sendButton" class="send-button img-button" onclick="sendEvent('chat:send')">
                    <img src="icons/send.svg" alt="Send">
                </button>
            </div>
        </div >
    `

        const chatInputArea = this.shadowRoot.querySelector('.chat-input-area')
        this.chatView = this.shadowRoot.querySelector('.chat-view')
        this.chatInput = this.shadowRoot.querySelector('#chatInput')
        this.cancelButton = this.shadowRoot.querySelector('.cancel-button')

        document.addEventListener('chat:last-message-update', e => this.onLastMessageUpdate(e.detail.sessionId))
        document.addEventListener('chat:send', e => this.sendMessageStreaming())
        document.addEventListener('storage:current-session', e => this.changeSession(e.detail))
        document.addEventListener('loading:generation-started', e => { if (this.chatSession.id == e.detail.sessionId) { this.cancelButton.classList.add('visible') } })
        document.addEventListener('loading:generation-stopped', e => { if (this.chatSession.id == e.detail.sessionId) { this.cancelButton.classList.remove('visible') } })
        document.addEventListener('chat:new-message', e => {
            if (this.chatSession.id == e.detail.sessionId) {
                this.appendMessage(e.detail)
            }
        })

        this.chatInput.addEventListener('input', () => {
            let isScrolledToBottom = this.chatView.scrollHeight - this.chatView.scrollTop <= (this.chatView.clientHeight + 15)
            this.chatInput.style.height = 'auto';
            const scrollHeight = this.chatInput.scrollHeight;
            const maxHeight = 150;

            this.chatInput.style.height = `${Math.min(scrollHeight, maxHeight)} px`;
            this.chatInput.style.overflowY = scrollHeight > maxHeight ? 'auto' : 'hidden';
            if (isScrolledToBottom) {
                this.scrollToBottom()
            }
        });

        this.chatInput.addEventListener('blur', () => {
            if (this.chatInput.value === '') {
                this.chatInput.style.height = 'auto';
                this.chatInput.style.overflowY = 'hidden';
            }
        });

        this.chatInput.addEventListener('keydown', (event) => {
            if (event.key === 'Enter' && !event.shiftKey) {
                event.preventDefault();
                this.sendMessageStreaming();
            }
        });

        const toolsCheckbox = chatInputArea.querySelector('ui-checkbox')
        toolsCheckbox.addEventListener('change', (e) => {
            const isChecked = e.target.checked
            this.toolsSelected = isChecked
        });

        this.cancelButton.addEventListener('click', e => { sendEvent('loading:generation-cancel', { sessionId: this.chatSession.id }) })
    }

    onLastMessageUpdate(sessionId) {
        if (this.chatSession && this.chatSession.id == sessionId && this.chatSession.messages && this.chatSession.messages.length > 0) {
            try {
                const messageElement = [...this.chatView.querySelectorAll('.message.assistant')].slice(-1)[0]
                const thinkSummary = messageElement.querySelector('.thinking-summary');
                const thinkContent = messageElement.querySelector('.thinking-content');
                const toolContent = messageElement.querySelector('.tool-content');
                const messageContent = messageElement.querySelector('.message-content')
                this.setAssistantMessageContent(messageContent, thinkContent, thinkSummary, toolContent, this.chatSession.messages[this.chatSession.messages.length - 1].text, this.chatSession.messages[this.chatSession.messages.length - 1].toolRequests)
            } catch {
                console.error(`Trying to update last message in chat but it doesnt exist, session: ${sessionId} `)
            }
        }
    }

    setAssistantMessageContent(messageElement, thinkElement, thinkSummary, toolContent, content, toolCalls) {
        try {
            let isScrolledToBottom = this.chatView.scrollHeight - this.chatView.scrollTop <= (this.chatView.clientHeight + 15)

            if (content.includes('<think>') && !content.includes('</think>')) {
                thinkSummary.classList.add('in-progress')
                content += '</think>'
            } else {
                thinkSummary.classList.remove('in-progress')
            }

            let thinkContent = content.match(/<think>([\s\S]*?)<\/think>/)
            if (thinkContent && thinkContent.length == 2) {
                let trimThinking = thinkContent[1].trim()
                if (trimThinking) {
                    thinkElement.textContent = thinkContent[1]
                    thinkElement.classList.remove('thinking-content-empty')
                }
            }
            let processedContent = content.replace(/<think>([\s\S]*?)<\/think>/g, '').trim();

            const htmlContent = marked.parse(processedContent, {
                gfm: true,
                breaks: true,
                mangle: false,
                headerIds: false
            });
            messageElement.innerHTML = htmlContent;

            this.applySyntaxHighlighting(messageElement);

            if (toolCalls && Array.isArray(toolCalls) && toolCalls.length > 0) {
                toolContent.classList.remove('tool-content-empty')
                toolContent.textContent = JSON.stringify(toolCalls)
            }

            if (isScrolledToBottom) {
                this.scrollToBottom()
            }

        } catch (e) {
            console.error(e)
            messageElement.textContent = content;
        }
    }

    applySyntaxHighlighting(element) {
        element.querySelectorAll('pre code').forEach((block) => {
            try {
                hljs.highlightElement(block);
            } catch { }
        });
    }

    scrollToBottom() {
        this.chatView.scrollTop = this.chatView.scrollHeight;
    }

    initAssisstantMessageElement(messageElement) {
        messageElement.innerHTML = `<div class="thinking-block">
            <div class="thinking-summary">💡 Thinking...</div>
            <div class="thinking-content thinking-content-empty"></div>
        </div>
        <div class="tool-block">
            <div class="tool-summary">🛠️ Tool call...</div>
            <div class="tool-content tool-content-empty"></div>
        </div>
        <div class="message-content"></div>`;

        const thinkBlock = messageElement.querySelector('.thinking-block');
        const thinkSummary = messageElement.querySelector('.thinking-summary');
        const thinkContent = messageElement.querySelector('.thinking-content');
        const toolBlock = messageElement.querySelector('.tool-block');
        const toolSummary = messageElement.querySelector('.tool-summary');
        const toolContent = messageElement.querySelector('.tool-content');
        const messageContent = messageElement.querySelector('.message-content');
        thinkSummary.addEventListener('click', () => {
            thinkBlock.classList.toggle('open');
        });
        toolSummary.addEventListener('click', () => {
            toolBlock.classList.toggle('open');
        });
        return { messageContent, thinkContent, thinkSummary, toolContent }
    }

    appendMessage(message) {
        const messageElement = document.createElement('div')
        messageElement.classList.add('message', message.origin)

        const messageInnerContent = document.createElement('div');
        messageInnerContent.classList.add('message-inner-content');

        if (message.origin === 'assistant') {
            const { messageContent, thinkContent, thinkSummary, toolContent } = this.initAssisstantMessageElement(messageInnerContent);
            this.setAssistantMessageContent(messageContent, thinkContent, thinkSummary, toolContent, message.text, message.toolRequests);
        } else {
            messageInnerContent.textContent = message.text;
        }
        messageElement.appendChild(messageInnerContent);

        const copyDeleteButtonsHTML = `<div class="copy-delete-buttons ${message.origin}">
            <button title="Copy" class="img-button" alt="Copy">📋</button>
            <button title="Delete" class="img-button"><img src="icons/delete.svg" alt="Delete"/></button>
        </div>`;
        messageElement.insertAdjacentHTML('beforeend', copyDeleteButtonsHTML);

        // Add event listeners for copy/delete (ensure you handle content extraction correctly)
        const buttons = messageElement.querySelectorAll('.copy-delete-buttons button');
        buttons[0].addEventListener('click', () => {
            // Smartly get content from messageInnerContent
            let contentToCopy = messageInnerContent.innerText || messageInnerContent.textContent;
            if (message.origin === 'assistant') {
                const mc = messageInnerContent.querySelector('.message-content');
                if (mc) contentToCopy = mc.innerText;
            }
            navigator.clipboard.writeText(contentToCopy)
        });

        this.chatView.appendChild(messageElement);
        this.scrollToBottom();
    }

    deleteMessage(messageElement) {
        // messageElement.remove();
    }

    async sendMessageStreaming() {
        sendEvent('sessions:touch')
        const messageText = this.chatInput.value.trim();
        if (!messageText) {
            return;
        }

        this.chatInput.value = '';
        this.chatInput.style.height = 'auto';
        this.chatInput.style.overflowY = 'hidden';
        this.chatInput.focus();

        this.scrollToBottom()
        const sessionId = this.chatSession.id

        if (this.toolsSelected) {
            apiToolChatStreaming(this.chatSession.id, messageText)
        } else {
            apiDirectChatStreaming(this.chatSession.id, messageText)
        }
    }

    changeSession(session) {
        if (session) {
            this.chatSession = session
            if (ongoingGenRequests.has(session.id)) {
                this.cancelButton.classList.add('visible')
            } else {
                this.cancelButton.classList.remove('visible')
            }
            if (session.messages && Array.isArray(session.messages)) {
                this.chatView.innerHTML = '';
                session.messages.forEach(message => {
                    this.appendMessage(message);
                });
                this.scrollToBottom();
            }
        } else {
            console.error('trying to change null session')
        }
    }

    cancelGeneration() {
        sendEvent('loading:generation-cancel', { sessionId: this.chatSession.id })
    }

}
customElements.define('chat-view', ChatView)