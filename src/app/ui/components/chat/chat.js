'use strict'

class ChatView extends HTMLElement {
    constructor() {
        super()

        this.chatSession = null

        const shadowRoot = this.attachShadow({ mode: 'open' })

        const styles = document.createElement('style')
        styles.innerHTML = `
        @import url('global.css');
        @import url('components/chat/chat.css');
        @import url('components/chat/syntax-theme.min.css');
        `
        shadowRoot.appendChild(styles)

        this.chatView = document.createElement('div')
        this.chatView.classList.add('chat-view')
        shadowRoot.appendChild(this.chatView)

        const chatInputArea = document.createElement('div')
        chatInputArea.classList.add('chat-input-area')
        chatInputArea.innerHTML = `
                    <textarea id="chatInput" class="chat-input" placeholder="Enter your message..." rows="1"></textarea>
                    <button id="sendButton" class="send-button" onclick="sendMessage('chat:send')">Send</button>
        `
        shadowRoot.appendChild(chatInputArea)
        this.chatInput = chatInputArea.querySelector('#chatInput')

        document.addEventListener('chat:last-message-update', e => this.onLastMessageUpdate(e.detail.sessionId))
        document.addEventListener('chat:send', e => this.sendMessageStreaming())
        document.addEventListener('chat:change-session', e => this.changeSession(e.detail.session))

        this.chatInput.addEventListener('input', () => {
            this.chatInput.style.height = 'auto';
            const scrollHeight = this.chatInput.scrollHeight;
            const maxHeight = 150;

            this.chatInput.style.height = `${Math.min(scrollHeight, maxHeight)}px`;
            this.chatInput.style.overflowY = scrollHeight > maxHeight ? 'auto' : 'hidden';
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
    }

    onLastMessageUpdate(sessionId) {
        if (this.chatSession && this.chatSession.id == sessionId && this.chatSession.messages && this.chatSession.messages.length > 0) {
            try {
                const messageElement = [...this.chatView.querySelectorAll('.message.assistant')].slice(-1)[0]
                const thinkSummary = messageElement.querySelector('.thinking-summary');
                const thinkContent = messageElement.querySelector('.thinking-content');
                const messageContent = messageElement.querySelector('.message-content')
                this.setAssistantMessageContent(messageContent, thinkContent, thinkSummary, this.chatSession.messages[this.chatSession.messages.length - 1].text)
            } catch {
                console.error(`Trying to update last message in chat but it doesnt exist, session: ${sessionId}`)
            }
        }
    }

    setAssistantMessageContent(messageElement, thinkElement, thinkSummary, content) {
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
            <div class="thinking-summary">🧠 Thinking...</div>
            <div class="thinking-content thinking-content-empty"></div>
        </div>
        <div class="message-content"></div>`;

        const thinkBlock = messageElement.querySelector('.thinking-block');
        const thinkSummary = messageElement.querySelector('.thinking-summary');
        const thinkContent = messageElement.querySelector('.thinking-content');
        const messageContent = messageElement.querySelector('.message-content');
        thinkSummary.addEventListener('click', () => {
            thinkBlock.classList.toggle('open');
        });
        return { messageContent, thinkContent, thinkSummary }
    }

    appendMessage(content, type) {
        const messageElement = document.createElement('div');
        messageElement.classList.add('message', type);

        if (type === 'assistant') {
            const { messageContent, thinkContent, thinkSummary } = this.initAssisstantMessageElement(messageElement)
            this.setAssistantMessageContent(messageContent, thinkContent, thinkSummary, content)
        } else {
            messageElement.textContent = content;
        }

        this.chatView.appendChild(messageElement);
        this.scrollToBottom();
    }

    async sendMessageStreaming() {
        sendEvent('sessions:touch')
        const messageText = this.chatInput.value.trim();
        if (!messageText) {
            return;
        }

        this.chatSession.messages.push({
            text: messageText,
            origin: 'user'
        })
        this.appendMessage(messageText, 'user');

        this.chatInput.value = '';
        this.chatInput.style.height = 'auto';
        this.chatInput.style.overflowY = 'hidden';
        this.chatInput.focus();

        const assistantMessageElement = document.createElement('div');
        assistantMessageElement.classList.add('message', 'assistant');
        this.initAssisstantMessageElement(assistantMessageElement)
        this.chatView.appendChild(assistantMessageElement);
        this.chatSession.messages.push({
            text: '',
            origin: 'assistant'
        })

        this.scrollToBottom()
        const sessionId = this.chatSession.id
        apiDirectChatStreaming(this.chatSession.id, messageText, chunk => updateLastMessage(sessionId, chunk))
    }

    changeSession(session) {
        if (session) {
            this.chatSession = session
            if (session.messages && Array.isArray(session.messages)) {
                this.chatView.innerHTML = '';
                session.messages.forEach(message => {
                    this.appendMessage(message.text, message.origin);
                });
                this.scrollToBottom();
            }
        } else {
            console.error('trying to change null session')
        }
    }

}
customElements.define('chat-view', ChatView)