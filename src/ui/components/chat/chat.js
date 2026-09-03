'use strict'

class ChatView extends HTMLElement {
    constructor() {
        super()

        this.chatSession = null
        this.toolsSelected = true
        this.searchQuery = ''
        this.searchTimer = null
        this.searchMatches = []
        this.activeSearchMatch = -1

        const shadowRoot = this.attachShadow({ mode: 'open' })

        this.shadowRoot.innerHTML = `
        <div class="find-panel" hidden>
            <input class="find-input" type="search" aria-label="Find in chat" placeholder="Find">
            <span class="find-count">0 matches</span>
            <button class="find-previous" title="Previous match" aria-label="Previous match">↑</button>
            <button class="find-next" title="Next match" aria-label="Next match">↓</button>
            <button class="find-close" title="Close" aria-label="Close">×</button>
        </div>
        <div class="chat-view">
        </div>
        <div class="chat-input-area">
            <button class="cancel-button">Stop</button >
            <div class="chat-input-container">
                <textarea id="chatInput" class="chat-input" placeholder="Enter your message..." rows="1"></textarea>
            </div>
            <div class="chat-button-container">
                <ui-checkbox class="tools-checkbox" label="Tools" ${this.toolsSelected ? 'checked' : ''}></ui-checkbox>
                <button id="sendButton" class="send-button img-button" alt="Send" onclick="sendEvent('chat:send')">&</button>
            </div>
        </div >
    `

        const chatInputArea = this.shadowRoot.querySelector('.chat-input-area')
        this.chatView = this.shadowRoot.querySelector('.chat-view')
        this.chatInput = this.shadowRoot.querySelector('#chatInput')
        this.cancelButton = this.shadowRoot.querySelector('.cancel-button')
        this.findPanel = this.shadowRoot.querySelector('.find-panel')
        this.findInput = this.shadowRoot.querySelector('.find-input')
        this.findCount = this.shadowRoot.querySelector('.find-count')

        document.addEventListener('chat:last-message-update', e => this.onLastMessageUpdate(e.detail.sessionId))
        document.addEventListener('chat:send', e => this.sendMessageStreaming())
        document.addEventListener('storage:current-session', e => this.changeSession(e.detail))
        document.addEventListener('loading:generation-started', e => { if (this.chatSession.id == e.detail.sessionId) { this.cancelButton.classList.add('visible') } })
        document.addEventListener('loading:generation-stopped', e => { if (this.chatSession.id == e.detail.sessionId) { this.cancelButton.classList.remove('visible') } })
        document.addEventListener('session:update', e => { if (this.chatSession.id == e.detail.session.id) { this.changeSession(e.detail.session) } })
        document.addEventListener('chat:new-message', e => {
            if (this.chatSession.id == e.detail.sessionId) {
                this.appendMessage(e.detail)
            }
        })
        document.addEventListener('keydown', e => this.onDocumentKeydown(e))

        this.findInput.addEventListener('input', () => this.queueSearch(this.findInput.value))
        this.findInput.addEventListener('keydown', e => {
            if (e.key === 'Escape') {
                e.preventDefault()
                this.closeSearch()
            } else if (e.key === 'Enter') {
                e.preventDefault()
                this.navigateSearch(e.shiftKey ? -1 : 1)
            }
        })
        this.shadowRoot.querySelector('.find-close').addEventListener('click', () => this.closeSearch())
        this.shadowRoot.querySelector('.find-previous').addEventListener('click', () => this.navigateSearch(-1))
        this.shadowRoot.querySelector('.find-next').addEventListener('click', () => this.navigateSearch(1))

        this.chatInput.addEventListener('input', () => {
            let isScrolledToBottom = this.chatView.scrollHeight - this.chatView.scrollTop <= (this.chatView.clientHeight + 15);
            this.chatInput.style.height = 'auto';
            const scrollHeight = this.chatInput.scrollHeight;
            const maxHeight = 150;

            this.chatInput.style.height = `${Math.min(scrollHeight, maxHeight)}px`;
            this.chatInput.style.overflowY = scrollHeight > maxHeight ? 'auto' : 'hidden';
            if (isScrolledToBottom) {
                this.scrollToBottom();
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
        this._initStyle()
    }

    async _initStyle() {
        this.shadowRoot.adoptedStyleSheets = [
            await loadCSS('global.css'),
            await loadCSS('components/chat/chat.css'),
            await loadCSS('components/chat/syntax-theme.min.css')
        ]
        this.scrollToBottom()
    }

    onLastMessageUpdate(sessionId) {
        if (this.chatSession && this.chatSession.id == sessionId && this.chatSession.messages && this.chatSession.messages.length > 0) {
            try {
                const messageElement = [...this.chatView.querySelectorAll('.message.assistant')].slice(-1)[0]
                const thinkSummary = messageElement.querySelector('.thinking-summary');
                const thinkContent = messageElement.querySelector('.thinking-content');
                const toolContent = messageElement.querySelector('.tool-content');
                const messageContent = messageElement.querySelector('.message-content')
                const message = this.chatSession.messages[this.chatSession.messages.length - 1]
                this.setAssistantMessageContent(messageContent, thinkContent, thinkSummary, toolContent, message.text, message.toolRequests)
                this.setResponseStatistics(messageElement.querySelector('.response-statistics'), message)
                this.reapplySearch()
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
            processedContent = processedContent.replace("<tool_call>", "```").replace("</tool_call>", "").trim();

            // --- MODIFIED: Wrap code blocks with header/footer and copy icon ---
            const htmlContent = marked.parse(processedContent, {
                gfm: true,
                breaks: true,
                mangle: false,
                headerIds: false,
                highlight: function (code, lang) {
                    // Let highlight.js handle highlighting
                    return hljs.highlightAuto(code, [lang]).value;
                }
            });

            // Wrap code blocks with header/footer
            const wrappedHtml = htmlContent.replace(
                /<pre><code( class="[^"]*")?>([\s\S]*?)<\/code><\/pre>/g,
                (match, cls, code) => {
                    // Unescape HTML entities for copying
                    const codeText = code.replace(/&lt;/g, "<").replace(/&gt;/g, ">").replace(/&amp;/g, "&");
                    return `
                    <div class="code-block-wrapper">
                        <div class="code-block-header">
                            <button class="copy-code-btn img-button" title="Copy code">7 <span>copy</span></button>
                        </div>
                        <pre><code${cls ? cls : ''}>${code}</code></pre>
                        <div class="code-block-footer">
                            <button class="copy-code-btn img-button" title="Copy code">7 <span>copy</span></button>
                        </div>
                    </div>
                    `;
                }
            );

            messageElement.innerHTML = wrappedHtml;

            // Add copy event listeners for all copy buttons in this message
            messageElement.querySelectorAll('.copy-code-btn').forEach(btn => {
                btn.addEventListener('click', (e) => {
                    const codeBlock = btn.closest('.code-block-wrapper').querySelector('pre code');
                    if (codeBlock) {
                        // Get plain text for copying
                        let codeText = codeBlock.innerText;
                        navigator.clipboard.writeText(codeText);
                    }
                });
            });

            messageElement.querySelectorAll('a').forEach(link => {
                let href = link.href
                link.setAttribute('data-href', href)
                link.href = ""
                link.addEventListener('click', (e) => {
                    e.preventDefault()
                    apiOpenLink(href)
                })
            })

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

    setResponseStatistics(element, message) {
        const outputTokens = message.outputTokens
        const elapsedMilliseconds = message.elapsedMilliseconds
        if (!element || !Number.isFinite(outputTokens) || outputTokens <= 0 || !Number.isFinite(elapsedMilliseconds) || elapsedMilliseconds <= 0) {
            if (element) element.hidden = true
            return
        }

        const seconds = Math.max(1, Math.floor(elapsedMilliseconds / 1000))
        const duration = seconds >= 60 ? `${Math.floor(seconds / 60)}m ${seconds % 60}s` : `${seconds}s`
        element.textContent = `${Math.floor(outputTokens / seconds)}t/s ${duration}`
        element.hidden = false
    }

    onDocumentKeydown(event) {
        if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'f') {
            const target = event.composedPath ? event.composedPath()[0] : event.target
            if (target instanceof HTMLInputElement || target instanceof HTMLTextAreaElement || target.isContentEditable) return
            event.preventDefault()
            this.openSearch()
        } else if (event.key === 'Escape' && !this.findPanel.hidden) {
            event.preventDefault()
            this.closeSearch()
        }
    }

    openSearch() {
        this.findPanel.hidden = false
        this.findInput.focus()
        this.findInput.select()
        this.reapplySearch()
    }

    closeSearch() {
        this.cancelSearchTimer()
        this.findPanel.hidden = true
        this.searchQuery = ''
        this.clearSearchHighlights()
        this.searchMatches = []
        this.activeSearchMatch = -1
        this.updateSearchControls()
    }

    cancelSearchTimer() {
        if (this.searchTimer !== null) {
            clearTimeout(this.searchTimer)
            this.searchTimer = null
        }
    }

    queueSearch(query) {
        this.cancelSearchTimer()
        this.searchQuery = query
        this.clearSearchHighlights()
        this.searchMatches = []
        this.activeSearchMatch = -1
        this.updateSearchControls()
        if (!query) return
        this.searchTimer = setTimeout(() => {
            this.searchTimer = null
            if (!this.findPanel.hidden && this.searchQuery === query) this.applySearch()
        }, 500)
    }

    clearSearchHighlights() {
        this.chatView.querySelectorAll('mark.chat-search-match').forEach(mark => {
            mark.replaceWith(document.createTextNode(mark.textContent))
        })
    }

    searchTextNodeGroups() {
        const groups = [
            ...this.chatView.querySelectorAll('.message.user .message-inner-content'),
            ...this.chatView.querySelectorAll('.message.assistant .message-content')
        ]
        return groups.map(root => {
            const nodes = []
            const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, {
                acceptNode: node => {
                    if (!node.nodeValue || node.parentElement.closest('.thinking-block, .tool-block, mark, button, .find-panel')) return NodeFilter.FILTER_REJECT
                    return NodeFilter.FILTER_ACCEPT
                }
            })
            let node
            while (node = walker.nextNode()) nodes.push(node)
            return nodes
        })
    }

    applySearch() {
        this.clearSearchHighlights()
        this.searchMatches = []
        this.activeSearchMatch = -1
        if (!this.searchQuery) {
            this.updateSearchControls()
            return
        }
        const query = this.searchQuery
        this.searchTextNodeGroups().forEach(nodes => {
            // Markdown and syntax highlighting can split one visible word across
            // adjacent text nodes (for example, "Lore" + "m"). Search the
            // rendered text as a stream, but keep the DOM structure intact.
            const text = nodes.map(node => node.nodeValue).join('')
            let offset = 0
            const matchRanges = []
            let index
            while ((index = text.indexOf(query, offset)) !== -1) {
                matchRanges.push({ start: index, end: index + query.length })
                offset = index + query.length
            }
            const matchMarks = matchRanges.map(() => null)
            let position = 0
            nodes.forEach(node => {
                const nodeStart = position
                const nodeEnd = position + node.nodeValue.length
                const fragments = []
                let localOffset = 0
                matchRanges.forEach((range, matchIndex) => {
                    const matchStart = Math.max(range.start, nodeStart)
                    const matchEnd = Math.min(range.end, nodeEnd)
                    if (matchStart < matchEnd) {
                        const localStart = matchStart - nodeStart
                        const localEnd = matchEnd - nodeStart
                        if (localStart > localOffset) fragments.push(document.createTextNode(node.nodeValue.slice(localOffset, localStart)))
                        const mark = document.createElement('mark')
                        mark.className = 'chat-search-match'
                        mark.dataset.searchMatch = String(this.searchMatches.length + matchIndex)
                        mark.textContent = node.nodeValue.slice(localStart, localEnd)
                        fragments.push(mark)
                        if (!matchMarks[matchIndex]) matchMarks[matchIndex] = mark
                        localOffset = localEnd
                    }
                })
                if (fragments.length) {
                    if (localOffset < node.nodeValue.length) fragments.push(document.createTextNode(node.nodeValue.slice(localOffset)))
                    node.replaceWith(...fragments)
                }
                position = nodeEnd
            })
            this.searchMatches.push(...matchMarks)
        })
        if (this.searchMatches.length) this.activeSearchMatch = 0
        this.updateSearchControls()
        this.setActiveSearchMatch()
    }

    reapplySearch() {
        if (!this.findPanel || this.findPanel.hidden || !this.searchQuery) return
        this.applySearch()
    }

    updateSearchControls() {
        this.findCount.textContent = `${this.searchMatches.length} match${this.searchMatches.length === 1 ? '' : 'es'}`
        const disabled = this.searchMatches.length === 0
        this.shadowRoot.querySelector('.find-previous').disabled = disabled
        this.shadowRoot.querySelector('.find-next').disabled = disabled
    }

    setActiveSearchMatch() {
        this.chatView.querySelectorAll('mark.chat-search-match').forEach(mark => {
            mark.classList.toggle('active', Number(mark.dataset.searchMatch) === this.activeSearchMatch)
        })
        if (this.activeSearchMatch >= 0) this.searchMatches[this.activeSearchMatch].scrollIntoView({ block: 'nearest' })
    }

    navigateSearch(direction) {
        if (!this.searchMatches.length) return
        this.activeSearchMatch = (this.activeSearchMatch + direction + this.searchMatches.length) % this.searchMatches.length
        this.setActiveSearchMatch()
    }

    scrollToBottom() {
        this.chatView.scrollTop = this.chatView.scrollHeight;
    }

    initAssisstantMessageElement(messageElement) {
        messageElement.innerHTML = `<div class="thinking-block">
            <div class="thinking-summary"><span class="icon">&#xe00d;</span> Thinking...</div>
            <div class="thinking-content thinking-content-empty"></div>
        </div>
        <div class="tool-block">
            <div class="tool-summary"><span class="icon">&#xe02a;</span> Tool call...</div>
            <div class="tool-content tool-content-empty"></div>
        </div>
        <div class="message-content"></div>
        <div class="response-statistics" hidden></div>`;

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
            this.setResponseStatistics(messageInnerContent.querySelector('.response-statistics'), message)
        } else if (message.origin === 'tool') {
            messageInnerContent.innerHTML = `
                <div class="tool-block">
                    <div class="tool-summary"><span class="icon">&#xe029;</span> Tool response</div>
                    <div class="tool-content">${message.text}</div>
                </div>
            `;
            const toolBlock = messageInnerContent.querySelector('.tool-block');
            const toolSummary = messageInnerContent.querySelector('.tool-summary');
            toolSummary.addEventListener('click', () => {
                toolBlock.classList.toggle('open');
            });
        } else {
            messageInnerContent.textContent = message.text;
        }
        messageElement.appendChild(messageInnerContent);

        if (message.origin == 'assistant' || message.origin == 'user') {
            const copyDeleteButtonsHTML = `<div class="copy-delete-buttons ${message.origin}">
                <button title="Copy" class="img-button" alt="Copy">7</button>
                <button title="Reload" class="img-button" alt="Generate again">Z</button>
                <button title="Delete" class="img-button" alt="Delete">&#xe053;</button>
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
            buttons[1].addEventListener('click', async () => {
                let messageToDelete = message

                if (message.origin != 'user') {
                    for (let i = 0; i <= this.chatSession.messages.length - 1; i++) {
                        let msg = this.chatSession.messages[i]
                        if (msg.id == message.id) {
                            for (let k = i - 1; k >= 0; k--) {
                                let backMsg = this.chatSession.messages[k]
                                if (backMsg.origin == 'user') {
                                    messageToDelete = backMsg
                                    break
                                }
                            }
                            break
                        }
                    }
                }
                const messageText = messageToDelete.text
                await apiTruncateSession(this.chatSession.id, messageToDelete.id)
                if (this.toolsSelected) {
                    apiToolChatStreaming(this.chatSession.id, messageText)
                } else {
                    apiDirectChatStreaming(this.chatSession.id, messageText)
                }
            });
            buttons[2].addEventListener('click', async () => {
                if (await confirmDialog("Delete this message?")) {
                    apiDeleteMessage(this.chatSession.id, message.id)
                }
            });

            if (message.origin != 'user') {
                const copyDeleteButtons = messageElement.querySelector('.copy-delete-buttons');
                if (copyDeleteButtons) {
                    // Reverse the order of the buttons
                    for (let i = copyDeleteButtons.children.length - 1; i >= 0; i--) {
                        copyDeleteButtons.appendChild(copyDeleteButtons.children[i]);
                    }
                }
            }
        }

        this.chatView.appendChild(messageElement);
        this.scrollToBottom();
        this.reapplySearch()
    }

    async sendMessageStreaming() {
        this.closeSearch()
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
                this.reapplySearch();
            }
        } else {
            console.error('trying to change null session')
        }
    }

    cancelGeneration() {
        sendEvent('loading:generation-cancel', { sessionId: this.chatSession.id })
    }

    // handleExternalLinkClick(event) {
    //     const linkElement = event.target.closest('a')

    //     if (linkElement && linkElement.href) {
    //         const url = new URL(linkElement.href)
    //         event.preventDefault()
    //         apiOpenLink(url.href)
    //     }
    // }

}
customElements.define('chat-view', ChatView)
