async function apiCreateSession() {
    try {
        const response = await fetch('/agent/sessions/new');
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
        const response = await fetch(`/agent/sessions/delete/${sessionId}`);
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
        const response = await fetch('/agent/sessions/list');
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

async function apiTruncateSession(sessionId, messageId) {
    try {
        const response = await fetch(`/agent/sessions/${sessionId}/truncate/${messageId}`)
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();

        return null
    } catch (error) {
        console.error("Failed to fetch sessions:", error);
        return null
    }
}

async function apiDeleteMessage(sessionId, messageId) {
    try {
        const response = await fetch(`/agent/sessions/${sessionId}/messages/delete/${messageId}`);
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }

        const data = await response.json();
    } catch (error) {
        console.error("Failed to delete session:", error);
    }
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

document.addEventListener('loading:generation-cancel', (e) => {
    sessionId = e.detail.sessionId
    if (ongoingGenRequests.has(sessionId)) {
        ongoingGenRequests.get(sessionId).abort()
        ongoingGenRequests.delete(sessionId)
    }
})

document.addEventListener('loading:generation-stopped', (e) => {
    sessionId = e.detail.sessionId
    if (ongoingGenRequests.has(sessionId)) {
        ongoingGenRequests.delete(sessionId)
    }
})

async function apiDirectChatStreaming(sessionId, message) {
    let controller = new AbortController()
    sendEvent('loading:generation-started', { sessionId: sessionId, controller: controller })

    try {
        const response = await fetch('/agent/directchat/stream', {
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

async function apiToolChatStreaming(sessionId, message) {
    let controller = new AbortController()
    sendEvent('loading:generation-started', { sessionId: sessionId, controller: controller })

    try {
        const response = await fetch('/agent/toolchat/stream', {
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
        const response = await fetch('/agent/models/list');
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

async function apiListProviders() {
    try {
        const response = await fetch('/agent/providers/list');
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();

        if (data && data.providers) {
            return data.providers
        } else {
            console.error("Invalid data structure received for providers:", data);
            return null
        }
    } catch (error) {
        console.error("Failed to list providers:", error);
        return null
    }
}

async function apiTestProvider(provider, signal) {
    try {
        const response = await fetch('/agent/provider/test', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(provider),
            signal
        });
        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
        const data = await response.json();
        return data ? data.response : null
    } catch (error) {
        if (error.name === 'AbortError') {
            return 'canceled'
        }
        console.error("Failed to test ai provider:", error);
        return null;
    }
}

async function apiUpdateProvider(provider) {
    try {
        const response = await fetch('/agent/provider/update', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(provider)
        });
        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
        const data = await response.json();
        return data
    } catch (error) {
        if (error.name === 'AbortError') {
            return 'canceled'
        }
        console.error("Failed to update provider:", error);
        return null;
    }
}

async function apiCreateProvider(provider) {
    try {
        const response = await fetch('/agent/provider/create', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(provider)
        });
        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
        const data = await response.json();
        return data;
    } catch (error) {
        if (error.name === 'AbortError') {
            return 'canceled'
        }
        console.error("Failed to create provider:", error);
        return null;
    }
}

async function apiDeleteProvider(id) {
    try {
        const response = await fetch(`/agent/provider/delete/${id}`);
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();
        return data;
    } catch (error) {
        console.error("Failed to delete provider:", error);
        return null
    }
}

async function apiListRoles() {
    try {
        const response = await fetch('/agent/roles/list');
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

async function apiCreateRole(role) {
    try {
        const response = await fetch('/agent/roles/create', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(role)
        });
        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
        const data = await response.json();
        return data.role;
    } catch (error) {
        console.error("Failed to create role:", error);
        return null;
    }
}

async function apiUpdateRole(role) {
    try {
        const response = await fetch('/agent/roles/update', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(role)
        });
        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
        const data = await response.json();
        return data.role;
    } catch (error) {
        console.error("Failed to update role:", error);
        return null;
    }
}

async function apiDeleteRole(id) {
    try {
        const response = await fetch(`/agent/roles/delete/${id}`);
        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
        const data = await response.json();
        return data;
    } catch (error) {
        console.error("Failed to delete role:", error);
        return null;
    }
}

async function apiListMCPServers() {
    try {
        const response = await fetch('/agent/mcp/list');
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

async function apiMCPTest(mcp, signal) {
    try {
        const response = await fetch('/agent/mcp/test', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(mcp),
            signal
        });
        if (!response.ok) throw new Error(`HTTP error! Status: ${response.status}`);
        const data = await response.json();
        return data ? data.response : null;
    } catch (error) {
        if (error.name === 'AbortError') {
            return 'canceled'
        }
        console.error("Failed to test mcp servers:", error);
        return null;
    }
}

async function apiMCPCreate(mcp) {
    try {
        const response = await fetch('/agent/mcp/create', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(mcp)
        });
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();

        if (data) {
            return data
        } else {
            console.error("Invalid data structure received for mcp servers:", data);
            return null
        }
    } catch (error) {
        console.error("Failed to list mcp servers:", error);
        return null
    }
}

async function apiMCPUpdate(mcp) {
    try {
        const response = await fetch('/agent/mcp/update', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(mcp)
        });
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();

        if (data) {
            return data
        } else {
            console.error("Invalid data structure received for mcp servers:", data);
            return null
        }
    } catch (error) {
        console.error("Failed to list mcp servers:", error);
        return null
    }
}

async function apiMCPDelete(id) {
    try {
        const response = await fetch(`/agent/mcp/delete/${id}`);
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();
    } catch (error) {
        console.error("Failed to delete mcp server:", error);
        return null
    }
}

async function apiAgentConnect() {
    var stream = new EventSource('/agent/sse');

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
            for (let i in Storage.sessions) {
                const session = Storage.sessions[i]
                if (session.id == parsedData.id) {
                    Storage.sessions[i] = parsedData
                    sendEvent('session:update', { session: parsedData })
                    break
                }
            }
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

    stream.addEventListener('mcp_list_update', function (event) {
        try {
            const parsedData = JSON.parse(event.data);
            sendEvent('mcps:reloaded', parsedData)
        } catch { }
    });

    stream.addEventListener('provider_list_update', function (event) {
        try {
            const parsedData = JSON.parse(event.data);
            sendEvent('providers:reloaded', parsedData)
        } catch { }
    });

    stream.addEventListener('role_list_update', function (event) {
        try {
            const parsedData = JSON.parse(event.data);
            sendEvent('roles:reloaded', parsedData)
        } catch { }
    });
}

async function apiOpenLink(url) {
    try {
        const response = await fetch('/agent/desktop/url/open', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                "url": url
            })
        });
        if (!response.ok) {
            throw new Error(`HTTP error! Status: ${response.status}`);
        }
        const data = await response.json();

        if (data) {
            return data
        }
    } catch (error) {
        console.error("Failed to open url:", error);
        return null
    }
}