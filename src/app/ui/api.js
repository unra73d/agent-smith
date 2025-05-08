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

async function apiDirectChatStreaming(sessionID, message, onMessage) {
    try {
        const response = await fetch('http://localhost:8008/agent/directchat/stream', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                "sessionID": sessionID,
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

async function apiListRoles(){
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