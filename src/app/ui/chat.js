const chatInput = document.getElementById('chatInput');
const chatView = document.getElementById('chatView');
const sendButton = document.getElementById('sendButton');
let chatSession = null

function applySyntaxHighlighting(element) {
    element.querySelectorAll('pre code').forEach((block) => {
        try {
            hljs.highlightElement(block);
        } catch{}
    });
}

function onLastMessageUpdate (sessionId){
    if(chatSession && chatSession.id == sessionId && chatSession.messages && chatSession.messages.length > 0){
        try{
            const messageElement = [...document.querySelectorAll('.message.assistant')].slice(-1)[0]
            const thinkSummary = messageElement.querySelector('.thinking-summary');
            const thinkContent = messageElement.querySelector('.thinking-content');
            const messageContent = messageElement.querySelector('.message-content')
            setAssistantMessageContent(messageContent, thinkContent, thinkSummary, chatSession.messages[chatSession.messages.length-1].text)
        } catch {
            console.error(`Trying to update last message in chat but it doesnt exist, session: ${sessionId}`)
        }
    }
}

document.addEventListener('chat:last-message-update', e=>onLastMessageUpdate(e.detail.sessionId))

function setAssistantMessageContent(messageElement, thinkElement, thinkSummary, content){
    try {
        let isScrolledToBottom = chatView.scrollHeight - chatView.scrollTop <= (chatView.clientHeight + 30)

        if(content.includes('<think>') && ! content.includes('</think>')){
            thinkSummary.classList.add('in-progress')
            content += '</think>'
        } else {
            thinkSummary.classList.remove('in-progress')
        }

        thinkContent = content.match(/<think>([\s\S]*?)<\/think>/)
        if (thinkContent && thinkContent.length == 2){
            trimThinking = thinkContent[1].trim()
            if (trimThinking){
                thinkElement.textContent = thinkContent[1]
                thinkElement.classList.remove('thinking-content-empty')
            }
        }
        processedContent = content.replace(/<think>([\s\S]*?)<\/think>/g, '').trim();

        const htmlContent = marked.parse(processedContent, {
            gfm: true,
            breaks: true,
            mangle: false,
            headerIds: false
        });
        messageElement.innerHTML = htmlContent;

        applySyntaxHighlighting(messageElement);
        
        if (isScrolledToBottom) {
            scrollToBottom()
        }

    } catch (e) {
        console.error(e)
        messageElement.textContent = content;
    }
}

function scrollToBottom() {
    chatView.scrollTop = chatView.scrollHeight;
}

function initAssisstantMessageElement(messageElement){
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
    return {messageContent, thinkContent, thinkSummary}
}

function appendMessage(content, type) {
    const messageElement = document.createElement('div');
    messageElement.classList.add('message', type);

    if (type === 'assistant') {
        const {messageContent, thinkContent, thinkSummary} = initAssisstantMessageElement(messageElement)
        setAssistantMessageContent(messageContent, thinkContent, thinkSummary, content)
    } else {
        messageElement.textContent = content;
    }

    chatView.appendChild(messageElement);
    scrollToBottom();
}

async function sendMessageStreaming() {
    sendEvent('sessions:touch')
    const messageText = chatInput.value.trim();
    if (!messageText) {
        return;
    }

    chatSession.messages.push({
        text: messageText,
        origin: 'user'
    })
    appendMessage(messageText, 'user');

    chatInput.value = '';
    chatInput.style.height = 'auto';
    chatInput.style.overflowY = 'hidden';
    chatInput.focus();

    const assistantMessageElement = document.createElement('div');
    assistantMessageElement.classList.add('message', 'assistant');
    const {messageContent, thinkContent, thinkSummary} = initAssisstantMessageElement(assistantMessageElement)
    chatView.appendChild(assistantMessageElement);
    chatSession.messages.push({
        text: '',
        origin: 'assistant'
    })

    scrollToBottom();

    apiDirectChatStreaming(currentSession.id, messageText, (chunk) => {
        // const isScrolledToBottom = chatView.scrollHeight - chatView.scrollTop <= chatView.clientHeight + 1;
        
        updateLastMessage(currentSession.id, chunk)
        
        // setAssistantMessageContent(messageContent, thinkContent, thinkSummary, content)
        // applySyntaxHighlighting(assistantMessageElement);
        
        // if (isScrolledToBottom) {
        //     scrollToBottom();
        // }
    });
}

sendButton.addEventListener('click', sendMessageStreaming);
const initialHeight = chatInput.scrollHeight;

chatInput.addEventListener('input', () => {
    chatInput.style.height = 'auto';
    const scrollHeight = chatInput.scrollHeight;
    const maxHeight = 150;

    chatInput.style.height = `${Math.min(scrollHeight, maxHeight)}px`;
    chatInput.style.overflowY = scrollHeight > maxHeight ? 'auto' : 'hidden';
});

chatInput.addEventListener('blur', () => {
    if (chatInput.value === '') {
        chatInput.style.height = 'auto';
        chatInput.style.overflowY = 'hidden';
    }
});

chatInput.addEventListener('keydown', async (event) => {
    if (event.key === 'Enter' && !event.shiftKey) {
        event.preventDefault();
        sendMessageStreaming();
    }
});

function chatChangeSession(session){
    chatSession = session
    if (session.messages && Array.isArray(session.messages)) {
        chatView.innerHTML = '';
        session.messages.forEach(message => {
            appendMessage(message.text, message.origin);
        });
        scrollToBottom();
    }
}