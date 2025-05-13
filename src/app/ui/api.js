async function apiCreateSession() {
    try {
        const response = await fetch('http://localhost:8008/agent/sessions/new');
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();

        if (data && data.session && data.session.id && data.session.summary && data.session.date) {
            return data.session
        } else {
            console.error("Invalid data structure received for new session:", data);
            return null
        }
    } catch (error) {
        console.error("Failed to create new session:", error);
        return null
    }
}

async function apiDeleteSession(sessionId) {
    try {
        const response = await fetch(`http://localhost:8008/agent/sessions/delete/${sessionId}`);
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }

        const data = await response.json();
    } catch (error) {
        console.error("Failed to delete session:", error);
    }
}

async function apiListSessions() {
    try {
        const response = await fetch('http://localhost:8008/agent/sessions/list');
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();

        if (data && data.sessions && Array.isArray(data.sessions)) {
            return data.sessions
        } else {
            return null
        }
    } catch (error) {
        console.error("Failed to fetch sessions:", error);
        return null
    }
}

function apiDirectChat() {

}

var ongoingGenRequests = new Map()
document.addEventListener('loading:generation-started', (e) => {
    sessionId = e.detail.sessionId
    if (ongoingGenRequests.has(sessionId)) {
        ongoingGenRequests.get(sessionId).abort()
        ongoingGenRequests.delete(sessionId)
    }
    ongoingGenRequests.set(sessionId, e.detail.controller)
})

document.addEventListener('loading:generation-stopped', (e) => {
    sessionId = e.detail.sessionId
    if (ongoingGenRequests.has(sessionId)) {
        ongoingGenRequests.get(sessionId).abort()
        ongoingGenRequests.delete(sessionId)
    }
})

async function apiDirectChatStreaming(sessionId, message, onMessage) {
    let controller = new AbortController()
    sendEvent('loading:generation-started', { sessionId: sessionId, controller: controller })

    try {
        const response = await fetch('http://localhost:8008/agent/directchat/stream', {
            signal: controller.signal,
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                "sessionID": sessionId,
                "modelID": getSelectedModelId(),
                "roleID": getSelectedRoleId(),
                "message": message
            })
        })
        const reader = response.body.pipeThrough(new TextDecoderStream()).getReader()
        while (true) {
            const { value, done } = await reader.read();
            if (done) break;
            onMessage(value)
        }
    } catch (error) {
        console.error("Failed to initiate streaming:", error);
    }
    sendEvent('loading:generation-stopped', { sessionId: sessionId })
}

async function apiToolChatStreaming(sessionId, message) {
    let controller = new AbortController()
    sendEvent('loading:generation-started', { sessionId: sessionId, controller: controller })

    try {
        const response = await fetch('http://localhost:8008/agent/toolchat/stream', {
            signal: controller.signal,
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                "sessionID": sessionId,
                "modelID": getSelectedModelId(),
                "roleID": getSelectedRoleId(),
                "message": message
            })
        })
        const reader = response.body.pipeThrough(new TextDecoderStream()).getReader()
        while (true) {
            const { value, done } = await reader.read();
            if (done) break;
        }
    } catch (error) {
        console.error("Failed to initiate streaming:", error);
    }
    sendEvent('loading:generation-stopped', { sessionId: sessionId })
}

async function apiListModels() {
    try {
        const response = await fetch('http://localhost:8008/agent/models/list');
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();

        if (data && data.models) {
            return data.models
        } else {
            console.error("Invalid data structure received for models:", data);
            return null
        }
    } catch (error) {
        console.error("Failed to list models:", error);
        return null
    }
}

async function apiListRoles() {
    try {
        const response = await fetch('http://localhost:8008/agent/roles/list');
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();

        if (data && data.roles) {
            return data.roles
        } else {
            console.error("Invalid data structure received for roles:", data);
            return null
        }
    } catch (error) {
        console.error("Failed to list roles:", error);
        return null
    }
}

async function apiListMCPServers() {
    try {
        const response = await fetch('http://localhost:8008/agent/mcp/list');
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();

        if (data && data.mcpServers) {
            return data.mcpServers
        } else {
            console.error("Invalid data structure received for mcp servers:", data);
            return null
        }
    } catch (error) {
        console.error("Failed to list mcp servers:", error);
        return null
    }
}

async function apiAgentConnect() {
    var stream = new EventSource('http://localhost:8008/agent/sse');

    stream.onopen = function (event) {
        console.log('Connection opened:', event);
    };

    stream.onmessage = function (event) {
        try {
            const parsedData = JSON.parse(event.data);
        } catch { }
    };

    stream.onerror = function (event) {
        console.error('Error event:', event);
        stream.close();
    };

    // Handle custom event types if needed
    stream.addEventListener('session_update', function (event) {
        try {
            const parsedData = JSON.parse(event.data);
        } catch { }
    });

    stream.addEventListener('new_message', function (event) {
        try {
            const parsedData = JSON.parse(event.data);
            for (let i in Storage.sessions) {
                const session = Storage.sessions[i]
                if (session.id == parsedData.sessionId) {
                    session.messages.push({
                        text: parsedData.message.text,
                        origin: parsedData.message.origin,
                        toolRequests: parsedData.message.toolRequests
                    })
                    break
                }
            }
            sendEvent('chat:new-message', { text: parsedData.message.text, origin: parsedData.message.origin, toolRequests: parsedData.message.toolRequests, sessionId: parsedData.sessionId })
        } catch (error) {
            console.error(error)
        }
    });

    stream.addEventListener('last_message_update', function (event) {
        try {
            const parsedData = JSON.parse(event.data);
            updateLastMessage(parsedData.sessionId, parsedData.message)
        } catch { }
    });

    stream.addEventListener('session_list_update', function (event) {
        try {
            const parsedData = JSON.parse(event.data);
        } catch { }
    });

    stream.addEventListener('model_list_update', function (event) {
        try {
            const parsedData = JSON.parse(event.data);
        } catch { }
    });

    stream.addEventListener('mcp_list_update', function (event) {
        try {
            const parsedData = JSON.parse(event.data);
        } catch { }
    });
}